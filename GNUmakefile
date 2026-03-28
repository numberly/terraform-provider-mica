BINARY_NAME=terraform-provider-flashblade
HOSTNAME=registry.terraform.io
NAMESPACE=soulkyu
TYPE=flashblade
VERSION=dev
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

.PHONY: build test testacc lint generate docs install default

default: build

build:
	go build -o $(BINARY_NAME)

test:
	go test ./internal/... -count=1

testacc:
	TF_ACC=1 go test ./... -count=1 -timeout 120m

lint:
	golangci-lint run ./...

generate:
	go generate ./...

docs:
	go generate ./...

install: build
	mkdir -p ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(TYPE)/$(VERSION)/$(OS_ARCH)
	cp $(BINARY_NAME) ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(TYPE)/$(VERSION)/$(OS_ARCH)/$(BINARY_NAME)
