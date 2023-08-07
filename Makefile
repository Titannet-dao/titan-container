SHELL=/usr/bin/env bash

all: build
.PHONY: all

unexport GOFLAGS

GOCC?=go

GOVERSION:=$(shell $(GOCC) version | tr ' ' '\n' | grep go1 | sed 's/^go//' | awk -F. '{printf "%d%03d%03d", $$1, $$2, $$3}')
ifeq ($(shell expr $(GOVERSION) \< 1017001), 1)
$(warning Your Golang version is go$(shell expr $(GOVERSION) / 1000000).$(shell expr $(GOVERSION) % 1000000 / 1000).$(shell expr $(GOVERSION) % 1000))
$(error Update Golang to version to at least 1.18)
endif

# git modules that need to be loaded
MODULES:=

CLEAN:=
BINS:=

ldflags=-X=github.com/Filecoin-Titan/titan-container/build.CurrentCommit=+git.$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"


manager: $(BUILD_DEPS)
	rm -f manager
	$(GOCC) build $(GOFLAGS) -o manager ./cmd/manager
.PHONY: manager

provider: $(BUILD_DEPS)
	rm -f provider
	$(GOCC) build $(GOFLAGS) -o provider ./cmd/provider
.PHONY: provider

api-gen:
	$(GOCC) run ./gen/api
	goimports -w api
.PHONY: api-gen

cfgdoc-gen:
	$(GOCC) run ./node/config/cfgdocgen > ./node/config/doc_gen.go

build: manager provider
.PHONY: build

install: install-manager install-provider

install-manager:
	install -C ./titan-manager /usr/local/bin/titan-manager

install-provider:
	install -C ./titan-provider /usr/local/bin/titan-provider
