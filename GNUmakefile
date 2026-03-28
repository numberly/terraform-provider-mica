SHELL := /bin/bash
export PATH := $(PATH):/usr/local/go/bin

BINARY_NAME=terraform-provider-flashblade
HOSTNAME=registry.terraform.io
NAMESPACE=numberly
TYPE=flashblade
VERSION=dev
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)
PLUGIN_DIR=~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(TYPE)/$(VERSION)/$(OS_ARCH)

.PHONY: build test testacc lint generate docs install dev-override clean default

default: build

build:
	go build -trimpath -o $(BINARY_NAME)

test:
	go test ./internal/... -count=1 -timeout 5m

testacc:
	TF_ACC=1 go test ./... -count=1 -timeout 120m

lint:
	golangci-lint run ./...

generate:
	go generate ./...

docs: generate

install: build
	mkdir -p $(PLUGIN_DIR)
	cp $(BINARY_NAME) $(PLUGIN_DIR)/terraform-provider-$(TYPE)

dev-override:
	@echo 'Add this to ~/.terraformrc:'
	@echo ''
	@echo 'provider_installation {'
	@echo '  dev_overrides {'
	@echo '    "$(NAMESPACE)/$(TYPE)" = "$(CURDIR)"'
	@echo '  }'
	@echo '  direct {}'
	@echo '}'
	@echo ''
	@echo 'Then run: make build'

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/
