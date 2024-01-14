# Contacts.app in THyMe

This repository demonstrates a modern-feeling web application in pure Golang using the ultra-minimal "THyMe" stack, ([Templ](https://templ.guide/), [HTMX](https://htmx.org/), [Minimux](https://github.com/meln5674/minimux)), with no custom JavaScript.

In addition to the application itself, it also demonstrates UI integration tests using [Ginkgo](https://onsi.github.io/ginkgo/) and [Biloba](https://onsi.github.io/biloba/)).

This application is more-or-less 1-to-1 with the example application from the book [Hypermedia Systems](https://hypermedia.systems/), (though currently incomplete) and the original Python version can be found [here](https://github.com/bigskysoftware/contact-app/tree/master).

## Pre-requisites

This project expects a [Go environment](https://go.dev/doc/install) to be installed and set up, along with a basic POSIX environment. All other tools be will installed by the build system as needed.

## Running

```
make run
# Or
make bin/app
bin/app [-addr localhost:8080] [-db db.json]
```

## Tests

```
make test
```

## Repository Contents

* model/model.go - Contact data model and fake database interface
* routes/routes.go - Constants and functions for building page paths
* templates/*.templ - Page and sub-page templates
* static/ - Static site data
* app.go - Entrypoint
* contacts_app_thyme_suite_test.go, app_test.go - Browser tests
* Makefile, make-env.yaml, make-env.Makefile - Build/Dependency management
