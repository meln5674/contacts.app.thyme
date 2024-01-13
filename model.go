package main

import (
	"strconv"
	"strings"
)

type ContactID int64

func ParseContactID(s string) (ContactID, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	return ContactID(id), err
}

type Contact struct {
	id     ContactID
	first  string
	last   string
	email  string
	phone  string
	errors map[string]string
}

type Contacts struct {
	contacts map[ContactID]Contact
}

func (c *Contacts) validate(contact *Contact) bool {
	errors := make(map[string]string, 4)
	if contact.email == "" {
		errors["email"] = "Email Required"
	}
	for _, other := range c.contacts {
		if other.email == contact.email && other.id != contact.id {
			errors["email"] = "Email Must Be Unique"
			break
		}
	}
	contact.errors = errors
	return len(contact.errors) == 0
}

func (c *Contacts) save(contact *Contact) bool {
	if !c.validate(contact) {
		return false
	}
	if contact.id == 0 {
		newID := ContactID(0)
		for _, other := range c.contacts {
			if other.id > newID {
				newID = other.id
			}
		}
		newID++
		contact.id = newID
	}
	c.contacts[contact.id] = *contact
	// TODO: DB
	return true
}

func (c *Contacts) search(q string) []Contact {
	results := make([]Contact, 0, len(c.contacts))
	for _, contact := range c.contacts {
		matches := strings.Contains(contact.first, q) ||
			strings.Contains(contact.last, q) ||
			strings.Contains(contact.email, q) ||
			strings.Contains(contact.phone, q)
		if matches {
			results = append(results, contact)
		}
	}
	return results
}

func (c *Contacts) all() []Contact {
	results := make([]Contact, 0, len(c.contacts))
	for _, contact := range c.contacts {
		results = append(results, contact)
	}
	return results
}

func (c *Contacts) find(id ContactID) (Contact, bool) {
	contact, ok := c.contacts[id]
	return contact, ok
}

func (c *Contacts) delete(id ContactID) {
	delete(c.contacts, id)
}
