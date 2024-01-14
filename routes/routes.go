package routes

import (
	"fmt"

	"github.com/meln5674/contacts.app.thyme/model"
)

const (
	IndexPath       = "/"
	ContactsPath    = "/contacts"
	NewContactPath  = "/contacts/new"
	ViewContactPath = "/contacts/([0-9]+)"
	EditContactPath = "/contacts/([0-9]+)/edit"
	// DeleteContactPath = "/contacts/([0-9]+)/delete"
	DeleteContactPath = ViewContactPath
	ContactEmailPath  = "/contacts/([0-9]+)/email"
	ContactsCountPath = "/contacts/count"
	StaticPath        = "/static/.*"
)

func ViewContactPathFor(id model.ContactID) string { return fmt.Sprintf("/contacts/%d", id) }
func EditContactPathFor(id model.ContactID) string { return fmt.Sprintf("/contacts/%d/edit", id) }

// func DeleteContactPathFor(id model.ContactID) string { return fmt.Sprintf("/contacts/%d/delete", id) }
func DeleteContactPathFor(id model.ContactID) string { return ViewContactPathFor(id) }
func ContactEmailPathFor(id model.ContactID) string  { return fmt.Sprintf("/contacts/%d/email", id) }
