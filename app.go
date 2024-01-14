package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/meln5674/minimux"

	"github.com/meln5674/contacts.app.thyme/model"
	"github.com/meln5674/contacts.app.thyme/routes"
	"github.com/meln5674/contacts.app.thyme/templates"
)

func main() {
	dbPath := flag.String("db", "db.json", "Path to Database JSON file")
	addr := flag.String("addr", "localhost:8080", "Address and port to listen on")
	flag.Parse()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	stop := make(chan struct{})
	go func() {
		<-signals
		signal.Stop(signals)
		close(stop)
	}()

	app := &App{Contacts: model.At(*dbPath)}
	err := app.LoadAndRun(*addr, stop)
	if err != nil {
		log.Fatal(err)
	}
}

type App struct {
	model.Contacts
}

func (a *App) LoadAndRun(addr string, stop chan struct{}) error {
	err := a.Contacts.Load()
	if err != nil {
		return fmt.Errorf("Could not load database: %#v", err)
	}
	srv := http.Server{
		Addr:    addr,
		Handler: a.Mux(),
	}
	go func() {
		<-stop
		srv.Close()
	}()
	log.Printf("Starting\n")
	err = srv.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func (a *App) Mux() *minimux.Mux {
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

func (a *App) contacts(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, formErr error) error {
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

func (a *App) contactsNewGet(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, _ error) error {
	w.WriteHeader(http.StatusOK)
	return templates.New(&model.Contact{}).Render(ctx, w)
}

func (a *App) contactsNew(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, formErr error) error {
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

func (a *App) contactsView(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
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

func (a *App) contactsEditGet(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
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

func (a *App) contactsEditPost(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
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

func (a *App) contactsDelete(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
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

func (a *App) contactsEmailGet(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
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

func (a *App) contactsCount(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, _ error) error {
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "(%d total Contacts)", a.Contacts.Count())
	return err
}
