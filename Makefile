# Some sensible Make defaults: https://tech.davis-hansson.com/p/make/
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
TOOLS_VERSIONS_FILE = $(PROJECT_DIR)/.tools_versions.yaml

export MISE_DATA_DIR = $(PROJECT_DIR)/bin/
MISE := $(shell which mise)
.PHONY: mise
mise:
	@mise -V >/dev/null || (echo "mise - https://github.com/jdx/mise - not found. Please install it." && exit 1)

GOLANGCI_LINT_VERSION = $(shell yq -ojson -r '.golangci-lint' < $(TOOLS_VERSIONS_FILE))
GOLANGCI_LINT = $(PROJECT_DIR)/bin/installs/golangci-lint/$(GOLANGCI_LINT_VERSION)/bin/golangci-lint
.PHONY: golangci-lint.download
golangci-lint.download: | mise ## Download golangci-lint locally if necessary.
	$(MISE) install -y -q golangci-lint@$(GOLANGCI_LINT_VERSION)

GORELEASER_VERSION = $(shell yq -ojson -r '.goreleaser' < $(TOOLS_VERSIONS_FILE))
GORELEASER = $(PROJECT_DIR)/bin/installs/goreleaser/$(GORELEASER_VERSION)/bin/goreleaser
.PHONY: goreleaser.download
goreleaser.download: | mise ## Download goreleaser locally if necessary.
	$(MISE) install -y -q goreleaser@$(GORELEASER_VERSION)

OAPI_CODEGEN_VERSION = $(shell yq -ojson -r '.oapi-codegen' < $(TOOLS_VERSIONS_FILE))
OAPI_CODEGEN = $(PROJECT_DIR)/bin/installs/oapi-codegen/$(OAPI_CODEGEN_VERSION)/bin/oapi-codegen
.PHONY: oapi-codegen.download
oapi-codegen.download: | mise ## Download oapi-codegen locally if necessary.
	$(MISE) install -y -q oapi-codegen@$(OAPI_CODEGEN_VERSION)

app/internal/api/gen.go: openapi.yaml | oapi-codegen.download
	$(OAPI_CODEGEN) --config openapi-config.yaml $<

.PHONY: clean
clean:
	@rm -rf app/internal/api/gen.go
	@rm -rf dist/
	@rm -rf bin/

.PHONY: all
all: check build test run

.PHONY: check
check: tidy fmt lint

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint: tidy golangci-lint

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: golangci-lint
golangci-lint: | golangci-lint.download
	$(GOLANGCI_LINT) run --verbose --config .golangci.yaml $(GOLANGCI_LINT_FLAGS)

build: | goreleaser.download
	$(GORELEASER) release --snapshot --clean

.PHONY: generate
generate: app/internal/api/gen.go
	go generate ./...

.PHONY: test
test:
	go test $$(go list ./... | grep -v 'e2e')
