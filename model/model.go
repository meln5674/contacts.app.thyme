package model

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
)

type ContactID int64

func ParseContactID(s string) (ContactID, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	return ContactID(id), err
}

type Contact struct {
	ID     ContactID `json:"id"`
	First  string    `json:"first"`
	Last   string    `json:"last"`
	Email  string    `json:"email"`
	Phone  string    `json:"phone"`
	Errors map[string]string
}

type Contacts struct {
	contacts map[ContactID]Contact
	path     string
}

func At(path string) Contacts {
	return Contacts{
		path:     path,
		contacts: make(map[ContactID]Contact),
	}
}

func (c *Contacts) Validate(contact *Contact) bool {
	errors := make(map[string]string, 4)
	if contact.Email == "" {
		errors["email"] = "Email Required"
	}
	for _, other := range c.contacts {
		if other.Email == contact.Email && other.ID != contact.ID {
			errors["email"] = "Email Must Be Unique"
			break
		}
	}
	contact.Errors = errors
	return len(contact.Errors) == 0
}

func (c *Contacts) Load() error {
	f, err := os.Open(c.path)
	if errors.Is(err, os.ErrNotExist) {
		c.contacts = make(map[ContactID]Contact)
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&c.contacts)
}

func (c *Contacts) Save(contact *Contact) (bool, error) {
	if !c.Validate(contact) {
		return false, nil
	}
	if contact.ID == 0 {
		newID := ContactID(0)
		for _, other := range c.contacts {
			if other.ID > newID {
				newID = other.ID
			}
		}
		newID++
		contact.ID = newID
	}
	c.contacts[contact.ID] = *contact

	f, err := os.Create(c.path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(&c.contacts)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Contacts) Search(q string) []Contact {
	results := make([]Contact, 0, len(c.contacts))
	for _, contact := range c.contacts {
		matches := strings.Contains(contact.First, q) ||
			strings.Contains(contact.Last, q) ||
			strings.Contains(contact.Email, q) ||
			strings.Contains(contact.Phone, q)
		if !matches {
			continue
		}
		results = append(results, contact)
	}
	return results
}

func (c *Contacts) All() []Contact {
	results := make([]Contact, 0, len(c.contacts))
	for _, contact := range c.contacts {
		results = append(results, contact)
	}
	return results
}

func (c *Contacts) Find(id ContactID) (Contact, bool) {
	contact, ok := c.contacts[id]
	return contact, ok
}

func (c *Contacts) Delete(id ContactID) error {
	delete(c.contacts, id)
	f, err := os.Create(c.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(&c.contacts)
}

func (c *Contacts) Count() int {
	return len(c.contacts)
}
