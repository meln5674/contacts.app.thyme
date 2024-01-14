
MAKE_ENV_BASE64 ?= $(shell command -v base64)
$(MAKE_ENV_BASE64): 
	stat $(MAKE_ENV_BASE64) >/dev/null
MAKE_ENV_CHMOD ?= $(shell command -v chmod)
$(MAKE_ENV_CHMOD): 
	stat $(MAKE_ENV_CHMOD) >/dev/null
MAKE_ENV_CURL ?= $(shell command -v curl)
$(MAKE_ENV_CURL): 
	stat $(MAKE_ENV_CURL) >/dev/null
MAKE_ENV_GO ?= $(shell command -v go)
$(MAKE_ENV_GO): 
	stat $(MAKE_ENV_GO) >/dev/null
MAKE_ENV_LN ?= $(shell command -v ln)
$(MAKE_ENV_LN): 
	stat $(MAKE_ENV_LN) >/dev/null
MAKE_ENV_MKDIR ?= $(shell command -v mkdir)
$(MAKE_ENV_MKDIR): 
	stat $(MAKE_ENV_MKDIR) >/dev/null
MAKE_ENV_RM ?= $(shell command -v rm)
$(MAKE_ENV_RM): 
	stat $(MAKE_ENV_RM) >/dev/null
MAKE_ENV_TAR ?= $(shell command -v tar)
$(MAKE_ENV_TAR): 
	stat $(MAKE_ENV_TAR) >/dev/null
MAKE_ENV_TOUCH ?= $(shell command -v touch)
$(MAKE_ENV_TOUCH): 
	stat $(MAKE_ENV_TOUCH) >/dev/null
MAKE_ENV_UNZIP ?= $(shell command -v unzip)
$(MAKE_ENV_UNZIP): 
	stat $(MAKE_ENV_UNZIP) >/dev/null
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	$(MAKE_ENV_MKDIR) -p $(LOCALBIN)
	$(MAKE_ENV_TOUCH) $(LOCALBIN)
$(LOCALBIN)/: $(LOCALBIN) 


TEMPL ?= $(LOCALBIN)/templ
$(TEMPL): $(MAKE_ENV_GO) 
	GOBIN=$(LOCALBIN)/.make-env/go/github.com/a-h/templ/cmd/templ/$(shell $(MAKE_ENV_GO) mod edit -print | grep github.com/a-h/templ | awk '{ print $$3 }') \
	$(MAKE_ENV_GO) install \
		github.com/a-h/templ/cmd/templ@$(shell $(MAKE_ENV_GO) mod edit -print | grep github.com/a-h/templ | awk '{ print $$3 }')
	$(MAKE_ENV_RM) -f $(TEMPL)
	$(MAKE_ENV_LN) -s $(LOCALBIN)/.make-env/go/github.com/a-h/templ/cmd/templ/$(shell $(MAKE_ENV_GO) mod edit -print | grep github.com/a-h/templ | awk '{ print $$3 }')/templ $(TEMPL)
.PHONY: templ
templ: $(TEMPL)

make-env.Makefile: make-env.yaml
	make-env --config 'make-env.yaml' --out 'make-env.Makefile'
