
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
	touch $(LOCALBIN)
$(LOCALBIN)/: $(LOCALBIN)


TEMPL ?= $(LOCALBIN)/templ
$(TEMPL):
	GOBIN=$(LOCALBIN)/.make-env/go/github.com/a-h/templ.cmd/templ/v0.2.513 go install github.com/a-h/templ/cmd/templ@v0.2.513
	rm -f $(TEMPL)
	ln -s $(LOCALBIN)/.make-env/go/github.com/a-h/templ.cmd/templ/v0.2.513/templ $(TEMPL)
.PHONY: templ
templ: $(TEMPL)

make-env.Makefile: make-env.yaml
	make-env --config 'make-env.yaml' --out 'make-env.Makefile'
