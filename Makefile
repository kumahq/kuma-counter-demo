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
SKAFFOLD = $(MISE) x skaffold -- skaffold
KUSTOMIZE = $(MISE) x kustomize -- kustomize

pkg/api/gen.go: openapi-config.yaml openapi.yaml
	$(OAPI_CODEGEN) --config openapi-config.yaml openapi.yaml

.PHONY: clean
clean:
	@rm -rf pkg/api/gen.go
	@rm -rf dist/
	@rm -rf k8s/

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

.PHONY: build
build:
	$(GORELEASER) release --snapshot --clean

k8s/base.yaml: kustomize/base/*
	@mkdir -p $(dir $@)
	kustomize build $(dir $<) > $@

k8s/%.yaml: kustomize/overlays/%/*
	@mkdir -p $(dir $@)
	$(KUSTOMIZE) build $(dir $<) > $@

k8s: k8s/base.yaml $(addprefix k8s/, $(addsuffix .yaml, $(shell find kustomize/overlays -mindepth 1 -type d | xargs -n 1 basename)))

.PHONY: generate
generate: pkg/api/gen.go k8s
	go generate ./...

.PHONY: test
test:
	go test $$(go list ./... | grep -v 'e2e')

.PHONY: skaffold/dev
skaffold/dev:
	$(SKAFFOLD) dev

.PHONY: demo/list
# List all demos
demo/list:
	@find kustomize -type f -name 'README.md' | sort | xargs -I {} -n1 -- bash -c 'head -1 {} && echo kubectl apply -k $$(dirname {})'

.PHONY: demo/add
# create a new demo
demo/add:
	@go run ./hack/new-entry/...
