all: bin/main

include make-env.Makefile

bin/main: main.go templates.templ.go
	go build -o bin/main ./

templates.templ.go: templates.templ $(TEMPL)
	$(TEMPL) generate

