all: bin/app

include make-env.Makefile

TEMPLATE_OUTS = \
	templates/edit_templ.go \
	templates/index_templ.go \
	templates/layout_templ.go \
	templates/new_templ.go \
	templates/rows_templ.go \
	templates/show_templ.go \


bin/app: app.go $(TEMPLATE_OUTS)
	go build -o bin/app ./

$(TEMPLATE_OUTS): $(TEMPL) $(shell find ./ -name '*.templ')
	$(TEMPL) generate || ( rm templates/*_templ.go; exit 1 )
	touch templates/*_templ.go

.PHONY: run
run: $(TEMPLATE_OUTS)
	go run ./

.PHONY: fmt
fmt: $(TEMPL)
	go fmt ./...
	$(TEMPL) fmt ./

.PHONY: vet
vet: $(TEMPLATE_OUTS)
	go vet ./...
