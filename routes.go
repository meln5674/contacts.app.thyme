package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/meln5674/minimux"
)

type app struct {
	Contacts
}

const (
	indexPath         = "/"
	contactsPath      = "/contacts"
	newContactPath    = "/contacts/new"
	viewContactPath   = "/contacts/([^/]+)"
	editContactPath   = "/contacts/([^/]+)/edit"
	deleteContactPath = "/contacts/([^/]+)/delete"
)

func viewContactPathFor(id ContactID) string   { return fmt.Sprintf("/contacts/%d", id) }
func editContactPathFor(id ContactID) string   { return fmt.Sprintf("/contacts/%d/edit", id) }
func deleteContactPathFor(id ContactID) string { return fmt.Sprintf("/contacts/%d/delete", id) }

func (a *app) Mux() *minimux.Mux {
	return &minimux.Mux{
		PreProcess: func(req *http.Request) (context.Context, func()) {
			fmt.Printf("%s %s %s\n", req.Method, req.URL, req.UserAgent())
			return minimux.CancelWhenDone(req)
		},
		PostProcess: func(ctx context.Context, req *http.Request, statusCode int, err error) {
			fmt.Printf("%s %s %s %d %v\n", req.Method, req.URL, req.UserAgent(), statusCode, err)
		},
		DefaultHandler: minimux.NotFound,
		Routes: []minimux.Route{
			minimux.
				LiteralPath(indexPath).
				WithMethods(http.MethodGet).
				IsHandledBy(minimux.RedirectingTo(contactsPath, http.StatusFound)),
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
				PathWithVars(editContactPath, "id").
				WithMethods(http.MethodGet).
				IsHandledByFunc(a.contactsEditGet),
			minimux.
				PathWithVars(editContactPath, "id").
				WithMethods(http.MethodPost).
				WithForm().
				IsHandledByFunc(a.contactsEditPost),
			minimux.
				PathWithVars(deleteContactPath, "id").
				WithMethods(http.MethodPost).
				IsHandledByFunc(a.contactsDelete),
		},
	}
}

func (a *app) contacts(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, formErr error) (int, error) {
	if formErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return http.StatusBadRequest, nil
	}
	search := req.Form.Get("q")
	var contacts []Contact
	if search != "" {
		contacts = a.Contacts.search(search)
	} else {
		contacts = a.Contacts.all()
	}
	w.WriteHeader(http.StatusOK)
	indexTpl(search, contacts).Render(ctx, w)
	return http.StatusOK, nil
}

func (a *app) contactsNewGet(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, _ error) (int, error) {
	w.WriteHeader(http.StatusOK)
	contactsNewTpl(&Contact{}).Render(ctx, w)
	return http.StatusOK, nil
}

func (a *app) contactsNew(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, formErr error) (int, error) {
	if formErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return http.StatusBadRequest, nil
	}
	c := Contact{
		first: req.Form.Get("first_name"),
		last:  req.Form.Get("last_name"),
		phone: req.Form.Get("phone"),
		email: req.Form.Get("email"),
	}
	if !a.Contacts.save(&c) {
		fmt.Printf("Failed to save %#v\n", c)
		w.WriteHeader(http.StatusOK)
		contactsNewTpl(&c).Render(ctx, w)
		return http.StatusOK, nil
	}
	fmt.Printf("Saved %#v\n", c)
	http.Redirect(w, req, contactsPath, http.StatusFound)
	return http.StatusFound, nil
}

func (a *app) contactsView(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) (int, error) {
	idStr := pathVars["id"]
	id, err := ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return http.StatusBadRequest, nil
	}
	contact, ok := a.Contacts.find(id)
	if !ok {
		fmt.Printf("No contact %d found\n", id)
		w.WriteHeader(http.StatusNotFound)
		return http.StatusNotFound, nil
	}
	fmt.Printf("Found contact %d %#v\n", id, contact)
	w.WriteHeader(http.StatusOK)
	contactsViewTpl(&contact).Render(ctx, w)
	return http.StatusOK, nil
}

func (a *app) contactsEditGet(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) (int, error) {
	idStr := pathVars["id"]
	id, err := ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return http.StatusBadRequest, nil
	}
	contact, ok := a.Contacts.find(id)
	if !ok {
		fmt.Printf("No contact %d found\n", id)
		w.WriteHeader(http.StatusNotFound)
		return http.StatusNotFound, nil
	}
	fmt.Printf("Found contact %d %#v\n", id, contact)
	w.WriteHeader(http.StatusOK)
	contactsEditTpl(&contact).Render(ctx, w)
	return http.StatusOK, nil
}

func (a *app) contactsEditPost(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) (int, error) {
	if formErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return http.StatusBadRequest, nil
	}
	idStr := pathVars["id"]
	id, err := ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return http.StatusBadRequest, nil
	}
	contact, ok := a.Contacts.find(id)
	if !ok {
		fmt.Printf("No contact %d found\n", id)
		w.WriteHeader(http.StatusNotFound)
		return http.StatusNotFound, nil
	}
	fmt.Printf("Found contact %d %#v\n", id, contact)
	contact = Contact{
		id:    contact.id,
		email: req.Form.Get("email"),
		first: req.Form.Get("first_name"),
		last:  req.Form.Get("last_name"),
		phone: req.Form.Get("phone"),
	}
	if !a.Contacts.save(&contact) {
		w.WriteHeader(http.StatusOK)
		contactsEditTpl(&contact).Render(ctx, w)
		return http.StatusOK, nil
	}
	http.Redirect(w, req, viewContactPathFor(id), http.StatusFound)
	return http.StatusFound, nil
}

func (a *app) contactsDelete(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) (int, error) {
	idStr := pathVars["id"]
	id, err := ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return http.StatusBadRequest, nil
	}
	contact, ok := a.Contacts.find(id)
	if !ok {
		fmt.Printf("No contact %d found\n", id)
		w.WriteHeader(http.StatusNotFound)
		return http.StatusNotFound, nil
	}
	fmt.Printf("Found contact %d %#v\n", id, contact)
	a.Contacts.delete(id)
	http.Redirect(w, req, contactsPath, http.StatusFound)
	return http.StatusFound, nil
}
