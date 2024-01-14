all: bin/main

include make-env.Makefile

TEMPLATE_OUTS=templates_templ.go

bin/main: main.go $(TEMPLATE_OUTS)
	go build -o bin/main ./

$(TEMPLATE_OUTS): $(TEMPL) $(shell find ./ -name '*.templ')
	$(TEMPL) generate || ( rm templates_templ.go; exit 1 )
	touch templates_templ.go

.PHONY: run
run: templates_templ.go
	go run ./

.PHONY: fmt
fmt: $(TEMPL)
	go fmt ./...
	$(TEMPL) fmt ./

.PHONY: vet
vet: $(TEMPLATE_OUTS)
	go vet ./...
