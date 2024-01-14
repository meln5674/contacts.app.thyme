package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/meln5674/minimux"

	"github.com/meln5674/hypermedia-systems-templ/model"
	"github.com/meln5674/hypermedia-systems-templ/routes"
	"github.com/meln5674/hypermedia-systems-templ/templates"
)

func main() {
	app := &app{Contacts: model.At("db.json")}
	err := app.Contacts.Load()
	if err != nil {
		fmt.Printf("Could not load database: %#v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Starting\n")
	http.ListenAndServe("localhost:8080", app.Mux())
}

type app struct {
	model.Contacts
}

func (a *app) Mux() *minimux.Mux {
	return &minimux.Mux{
		PreProcess:     minimux.PreProcessorChain(minimux.CancelWhenDone, minimux.LogPendingRequest(os.Stdout)),
		PostProcess:    minimux.LogCompletedRequest(os.Stdout),
		DefaultHandler: minimux.NotFound,
		Routes: []minimux.Route{
			minimux.
				PathPattern(routes.StaticPath).
				WithMethods(http.MethodGet).
				IsHandledBy(minimux.Simple(http.FileServer(http.Dir("./")))),
			minimux.
				LiteralPath(routes.IndexPath).
				WithMethods(http.MethodGet).
				IsHandledBy(minimux.RedirectingTo(routes.ContactsPath, http.StatusPermanentRedirect)),
			minimux.
				LiteralPath(routes.ContactsPath).
				WithMethods(http.MethodGet).
				WithForm().
				IsHandledByFunc(a.contacts),
			minimux.
				LiteralPath(routes.NewContactPath).
				WithMethods(http.MethodGet).
				IsHandledByFunc(a.contactsNewGet),
			minimux.
				LiteralPath(routes.NewContactPath).
				WithMethods(http.MethodPost).
				WithForm().
				IsHandledByFunc(a.contactsNew),
			minimux.
				PathWithVars(routes.ViewContactPath, "id").
				WithMethods(http.MethodGet).
				IsHandledByFunc(a.contactsView),

			minimux.
				PathWithVars(routes.EditContactPath, "id").
				WithMethods(http.MethodGet).
				IsHandledByFunc(a.contactsEditGet),
			minimux.
				PathWithVars(routes.EditContactPath, "id").
				WithMethods(http.MethodPost).
				WithForm().
				IsHandledByFunc(a.contactsEditPost),
			minimux.
				PathWithVars(routes.DeleteContactPath, "id").
				// WithMethods(http.MethodPost).
				WithMethods(http.MethodDelete).
				IsHandledByFunc(a.contactsDelete),
			minimux.
				PathWithVars(routes.ContactEmailPath, "id").
				WithMethods(http.MethodGet).
				WithForm().
				IsHandledByFunc(a.contactsEmailGet),
			minimux.
				LiteralPath(routes.ContactsCountPath).
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
	var contacts []model.Contact
	if search != "" {
		contacts = a.Contacts.Search(search)
	} else {
		contacts = a.Contacts.All()
	}
	w.WriteHeader(http.StatusOK)
	if req.Header.Get("HX-Trigger") == "search" {
		return templates.Rows(contacts).Render(ctx, w)
	}
	return templates.Index(search, page, contacts).Render(ctx, w)
}

func (a *app) contactsNewGet(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, _ error) error {
	w.WriteHeader(http.StatusOK)
	return templates.New(&model.Contact{}).Render(ctx, w)
}

func (a *app) contactsNew(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, formErr error) error {
	if formErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	c := model.Contact{
		First: req.Form.Get("first_name"),
		Last:  req.Form.Get("last_name"),
		Phone: req.Form.Get("phone"),
		Email: req.Form.Get("email"),
	}
	saved, err := a.Contacts.Save(&c)
	if err != nil {
		fmt.Printf("Failed to save %#v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	if !saved {
		fmt.Printf("Failed to save %#v\n", c)
		w.WriteHeader(http.StatusOK)
		return templates.New(&c).Render(ctx, w)
	}
	fmt.Printf("Saved %#v\n", c)
	http.Redirect(w, req, routes.ContactsPath, http.StatusSeeOther)
	return nil
}

func (a *app) contactsView(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
	idStr := pathVars["id"]
	id, err := model.ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	contact, ok := a.Contacts.Find(id)
	if !ok {
		fmt.Printf("No contact %d found\n", id)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	fmt.Printf("Found contact %d %#v\n", id, contact)
	w.WriteHeader(http.StatusOK)
	return templates.Show(&contact).Render(ctx, w)
}

func (a *app) contactsEditGet(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
	idStr := pathVars["id"]
	id, err := model.ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	contact, ok := a.Contacts.Find(id)
	if !ok {
		fmt.Printf("No contact %d found\n", id)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	fmt.Printf("Found contact %d %#v\n", id, contact)
	w.WriteHeader(http.StatusOK)
	return templates.Edit(&contact).Render(ctx, w)
}

func (a *app) contactsEditPost(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
	if formErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	idStr := pathVars["id"]
	id, err := model.ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	contact, ok := a.Contacts.Find(id)
	if !ok {
		fmt.Printf("No contact %d found\n", id)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	fmt.Printf("Found contact %d %#v\n", id, contact)
	contact = model.Contact{
		ID:    contact.ID,
		Email: req.Form.Get("email"),
		First: req.Form.Get("first_name"),
		Last:  req.Form.Get("last_name"),
		Phone: req.Form.Get("phone"),
	}
	saved, err := a.Contacts.Save(&contact)
	if err != nil {
		fmt.Printf("Failed to save %#v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	if !saved {
		w.WriteHeader(http.StatusOK)
		return templates.Edit(&contact).Render(ctx, w)
	}
	http.Redirect(w, req, routes.ViewContactPathFor(id), http.StatusSeeOther)
	return nil
}

func (a *app) contactsDelete(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
	idStr := pathVars["id"]
	id, err := model.ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	contact, ok := a.Contacts.Find(id)
	if !ok {
		fmt.Printf("No contact %d found\n", id)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	fmt.Printf("Found contact %d %#v\n", id, contact)
	err = a.Contacts.Delete(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	w.WriteHeader(http.StatusOK)
	if req.Header.Get("HX-Trigger") == "delete-btn" {
		http.Redirect(w, req, routes.ContactsPath, http.StatusSeeOther)
		return nil
	}
	return nil
}

func (a *app) contactsEmailGet(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
	idStr := pathVars["id"]
	id, err := model.ParseContactID(idStr)
	if err != nil {
		fmt.Printf("%s is not a valid id\n", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	contact, _ := a.Contacts.Find(id)
	contact.Email = req.Form.Get("email")
	a.Contacts.Validate(&contact)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(contact.Errors["email"]))
	return nil
}

func (a *app) contactsCount(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "(%d total Contacts)", a.Contacts.Count())
	return err
}
