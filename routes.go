package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/meln5674/minimux"
)

type app struct {
	Contacts
}

const (
	indexPath         = "/"
	contactsPath      = "/contacts"
	newContactPath    = "/contacts/new"
	viewContactPath   = "/contacts/([0-9]+)"
	editContactPath   = "/contacts/([0-9]+)/edit"
	deleteContactPath = "/contacts/([0-9]+)/delete"
	contactEmailPath  = "/contacts/([0-9]+)/email"
	contactsCountPath = "/contacts/count"
	staticPath        = "/static/.*"
)

func viewContactPathFor(id ContactID) string   { return fmt.Sprintf("/contacts/%d", id) }
func editContactPathFor(id ContactID) string   { return fmt.Sprintf("/contacts/%d/edit", id) }
func deleteContactPathFor(id ContactID) string { return fmt.Sprintf("/contacts/%d/delete", id) }

func (a *app) Mux() *minimux.Mux {
	return &minimux.Mux{
		PreProcess:     minimux.PreProcessorChain(minimux.CancelWhenDone, minimux.LogPendingRequest(os.Stdout)),
		PostProcess:    minimux.LogCompletedRequest(os.Stdout),
		DefaultHandler: minimux.NotFound,
		Routes: []minimux.Route{
			minimux.
				PathPattern(staticPath).
				WithMethods(http.MethodGet).
				IsHandledBy(minimux.Simple(http.FileServer(http.Dir("./")))),
			minimux.
				LiteralPath(indexPath).
				WithMethods(http.MethodGet).
				IsHandledBy(minimux.RedirectingTo(contactsPath, http.StatusPermanentRedirect)),
			minimux.
				LiteralPath(contactsPath).
				WithMethods(http.MethodGet).
				WithForm().
				IsHandledByFunc(a.contacts),
			minimux.
				LiteralPath(newContactPath).
				WithMethods(http.MethodGet).
				IsHandledByFunc(a.contactsNewGet),
			minimux.
				LiteralPath(newContactPath).
				WithMethods(http.MethodPost).
				WithForm().
				IsHandledByFunc(a.contactsNew),
			minimux.
				PathWithVars(viewContactPath, "id").
				WithMethods(http.MethodGet).
				IsHandledByFunc(a.contactsView),
			minimux.
				PathWithVars(viewContactPath, "id").
				WithMethods(http.MethodDelete).
				IsHandledByFunc(a.contactsDelete),
			minimux.
				PathWithVars(editContactPath, "id").
				WithMethods(http.MethodGet).
				IsHandledByFunc(a.contactsEditGet),
			minimux.
				PathWithVars(editContactPath, "id").
				WithMethods(http.MethodPost).
				WithForm().
				IsHandledByFunc(a.contactsEditPost),
			/*
				minimux.
					PathWithVars(deleteContactPath, "id").
					WithMethods(http.MethodPost).
					IsHandledByFunc(a.contactsDelete),
			*/
			minimux.
				PathWithVars(contactEmailPath, "id").
				WithMethods(http.MethodGet).
				WithForm().
				IsHandledByFunc(a.contactsEmailGet),
			minimux.
				LiteralPath(contactsCountPath).
				WithMethods(http.MethodGet).
				IsHandledByFunc(a.contactsCount),
		},
	}
}

func (a *app) contacts(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, formErr error) error {
	var err error
	if formErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	search := req.Form.Get("q")
	page := int64(1)
	pageStr := req.Form.Get("page")
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}
	}
	var contacts []Contact
	if search != "" {
		contacts = a.Contacts.search(search)
	} else {
		contacts = a.Contacts.all()
	}
	w.WriteHeader(http.StatusOK)
	if req.Header.Get("HX-Trigger") == "search" {
		return rowsTpl(contacts).Render(ctx, w)
	}
	return indexTpl(search, page, contacts).Render(ctx, w)
}

func (a *app) contactsNewGet(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, _ error) error {
	w.WriteHeader(http.StatusOK)
	return contactsNewTpl(&Contact{}).Render(ctx, w)
}

func (a *app) contactsNew(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, formErr error) error {
	if formErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	c := Contact{
		First: req.Form.Get("first_name"),
		Last:  req.Form.Get("last_name"),
		Phone: req.Form.Get("phone"),
		Email: req.Form.Get("email"),
	}
	saved, err := a.Contacts.save(&c)
	if err != nil {
		fmt.Printf("Failed to save %#v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	if !saved {
		fmt.Printf("Failed to save %#v\n", c)
		w.WriteHeader(http.StatusOK)
		return contactsNewTpl(&c).Render(ctx, w)
	}
	fmt.Printf("Saved %#v\n", c)
	http.Redirect(w, req, contactsPath, http.StatusSeeOther)
	return nil
}

func (a *app) contactsView(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
	idStr := pathVars["id"]
	id, err := ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	contact, ok := a.Contacts.find(id)
	if !ok {
		fmt.Printf("No contact %d found\n", id)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	fmt.Printf("Found contact %d %#v\n", id, contact)
	w.WriteHeader(http.StatusOK)
	return contactsViewTpl(&contact).Render(ctx, w)
}

func (a *app) contactsEditGet(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
	idStr := pathVars["id"]
	id, err := ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	contact, ok := a.Contacts.find(id)
	if !ok {
		fmt.Printf("No contact %d found\n", id)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	fmt.Printf("Found contact %d %#v\n", id, contact)
	w.WriteHeader(http.StatusOK)
	return contactsEditTpl(&contact).Render(ctx, w)
}

func (a *app) contactsEditPost(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
	if formErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	idStr := pathVars["id"]
	id, err := ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	contact, ok := a.Contacts.find(id)
	if !ok {
		fmt.Printf("No contact %d found\n", id)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	fmt.Printf("Found contact %d %#v\n", id, contact)
	contact = Contact{
		ID:    contact.ID,
		Email: req.Form.Get("email"),
		First: req.Form.Get("first_name"),
		Last:  req.Form.Get("last_name"),
		Phone: req.Form.Get("phone"),
	}
	saved, err := a.Contacts.save(&contact)
	if err != nil {
		fmt.Printf("Failed to save %#v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	if !saved {
		w.WriteHeader(http.StatusOK)
		return contactsEditTpl(&contact).Render(ctx, w)
	}
	http.Redirect(w, req, viewContactPathFor(id), http.StatusSeeOther)
	return nil
}

func (a *app) contactsDelete(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
	idStr := pathVars["id"]
	id, err := ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	contact, ok := a.Contacts.find(id)
	if !ok {
		fmt.Printf("No contact %d found\n", id)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	fmt.Printf("Found contact %d %#v\n", id, contact)
	err = a.Contacts.delete(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	w.WriteHeader(http.StatusOK)
	if req.Header.Get("HX-Trigger") == "delete-btn" {
		http.Redirect(w, req, contactsPath, http.StatusSeeOther)
		return nil
	}
	return nil
}

func (a *app) contactsEmailGet(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
	idStr := pathVars["id"]
	id, err := ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	contact, _ := a.Contacts.find(id)
	contact.Email = req.Form.Get("email")
	a.Contacts.validate(&contact)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(contact.errors["email"]))
	return nil
}

func (a *app) contactsCount(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "(%d total Contacts)", a.Contacts.count())
	return err
}
