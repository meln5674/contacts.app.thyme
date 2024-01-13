package main

import (
	"fmt"
	"net/http"
)

func main() {
	app := &app{
		Contacts: Contacts{
			contacts: map[ContactID]Contact{},
		},
	}
	fmt.Printf("Starting\n")
	http.ListenAndServe("localhost:8080", app.Mux())
}
