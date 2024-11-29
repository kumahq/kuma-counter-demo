# Some sensible Make defaults: https://tech.davis-hansson.com/p/make/
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

ifeq (,$(shell which mise))
$(error "mise - https://github.com/jdx/mise - not found. Please install it.")
endif
MISE := $(shell which mise) -y

GOLANGCI_LINT = $(MISE) x golangci-lint -- golangci-lint
GORELEASER = $(MISE) x goreleaser -- goreleaser
OAPI_CODEGEN = $(MISE) x oapi-codegen -- oapi-codegen

pkg/api/gen.go: openapi-config.yaml openapi.yaml
	$(OAPI_CODEGEN) --config openapi-config.yaml openapi.yaml

.PHONY: clean
clean:
	@rm -rf pkg/api/gen.go
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
golangci-lint:
	$(GOLANGCI_LINT) run --verbose --config .golangci.yaml $(GOLANGCI_LINT_FLAGS)

build:
	$(GORELEASER) release --snapshot --clean

.PHONY: generate
generate: pkg/api/gen.go
	go generate ./...

.PHONY: test
test:
	go test $$(go list ./... | grep -v 'e2e')
