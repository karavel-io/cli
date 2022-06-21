PROJ_DIR     := $(CURDIR)
BINDIR       := $(PROJ_DIR)/bin
BINNAME      ?= karavel
INSTALL_PATH ?= /usr/local/bin
SHELL        = /usr/bin/env bash

SRC := $(shell find . -type f -name '*.go' -print) go.mod go.sum

GOBIN         = $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN         = $(shell go env GOPATH)/bin
endif
PKGS          = $(PROJ_DIR)/...

VERSION    = $(shell cat $(PROJ_DIR)/VERSION)

LDFLAGS    += -X 'github.com/karavel-io/cli/internal/version.version=${VERSION}'
LDFLAGS    += $(EXT_LDFLAGS)

.PHONY: all
all: build

.PHONY: addlicense
addlicense:
	addlicense -c "The Karavel Project" -l apache .

.PHONY: build
build: fmt vet build-simple

.PHONY: build-simple
build-simple: $(BINDIR)/$(BINNAME)

$(BINDIR)/$(BINNAME): $(SRC)
	GO111MODULE=on go build -ldflags "$(LDFLAGS)" -o $(BINDIR)/$(BINNAME) $(PROJ_DIR)/cmd/karavel

.PHONY: install
install: build
	@install $(BINDIR)/$(BINNAME) $(INSTALL_PATH)/$(BINNAME)

.PHONY: clean
clean:
	rm -rf $(BINDIR)

.PHONY: test
test:
	go test $(PKGS)

.PHONY: fmt
fmt: addlicense
	go fmt $(PKGS)

.PHONY: vet
vet:
	go vet $(PKGS)

.PHONY: update-deps
update-deps:
	go get -u ./...
	go mod tidy
