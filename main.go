package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	app := &app{
		Contacts: Contacts{
			contacts: map[ContactID]Contact{},
			path:     "db.json",
		},
	}
	err := app.Contacts.load()
	if err != nil {
		fmt.Printf("Could not load database: %#v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Starting\n")
	http.ListenAndServe("localhost:8080", app.Mux())
}
