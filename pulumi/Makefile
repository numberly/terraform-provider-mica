PROJECT_NAME := pulumi-flashblade
PROVIDER_PKG := github.com/numberly/opentofu-provider-flashblade/pulumi/provider
VERSION := $(shell git describe --tags --dirty --always 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X $(PROVIDER_PKG)/pkg/version.Version=$(VERSION)"

PROVIDER_DIR := provider
TFGEN_BIN   := $(PROVIDER_DIR)/cmd/pulumi-tfgen-flashblade/pulumi-tfgen-flashblade
RUNTIME_BIN := $(PROVIDER_DIR)/cmd/pulumi-resource-flashblade/pulumi-resource-flashblade
SCHEMA_DIR  := $(PROVIDER_DIR)/cmd/pulumi-resource-flashblade

.PHONY: version tfgen_build tfgen provider test clean generate_python generate_go test_python_sdk test_go_sdk

version:
	@echo $(VERSION)

# Build tfgen binary (no ldflags version injection — tfgen is a build tool).
tfgen_build:
	cd $(PROVIDER_DIR) && go build -o cmd/pulumi-tfgen-flashblade/pulumi-tfgen-flashblade ./cmd/pulumi-tfgen-flashblade

# Generate schema artifacts. tfgen's `schema` subcommand emits schema.json + bridge-metadata.json
# into the --out directory. schema-embed.json is derived via jq minification.
tfgen: tfgen_build
	$(TFGEN_BIN) schema --out $(SCHEMA_DIR)
	@test -s $(SCHEMA_DIR)/schema.json          || (echo "schema.json empty" && exit 1)
	@test -s $(SCHEMA_DIR)/bridge-metadata.json || (echo "bridge-metadata.json empty" && exit 1)
	@jq -c . $(SCHEMA_DIR)/schema.json > $(SCHEMA_DIR)/schema-embed.json
	@test -s $(SCHEMA_DIR)/schema-embed.json    || (echo "schema-embed.json empty" && exit 1)

# Build runtime provider binary with ldflags-injected VERSION and real embedded schema.
provider: tfgen
	cd $(PROVIDER_DIR) && go build $(LDFLAGS) -o cmd/pulumi-resource-flashblade/pulumi-resource-flashblade ./cmd/pulumi-resource-flashblade

# SDK Generation Scope Boundary (SDK-04):
# Only Python and Go SDKs are generated. NodeJS, .NET, and Java are explicitly
# out of scope for this provider.

PYTHON_SDK_DIR := sdk/python
GO_SDK_DIR     := sdk/go

# Generate Python SDK from committed schema.
generate_python: tfgen
	@echo "Generating Python SDK..."
	@rm -rf $(PYTHON_SDK_DIR)
	@mkdir -p $(PYTHON_SDK_DIR)
	$(TFGEN_BIN) python --out $(PYTHON_SDK_DIR) --skip-docs --skip-examples
	@echo "Python SDK generated in $(PYTHON_SDK_DIR)"

# Generate Go SDK from committed schema. Preserves existing go.mod.
# Post-generation: patch import paths to include /pulumi/ prefix.
generate_go: tfgen
	@echo "Generating Go SDK..."
	@rm -rf $(GO_SDK_DIR)/*.go $(GO_SDK_DIR)/*/
	@cp $(GO_SDK_DIR)/go.mod $(GO_SDK_DIR)/go.mod.bak
	$(TFGEN_BIN) go --out $(GO_SDK_DIR) --skip-docs --skip-examples
	@find $(GO_SDK_DIR) -name "*.go" -exec sed -i \
		's|github.com/numberly/opentofu-provider-flashblade/sdk/go|github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go|g' {} +
	@mv $(GO_SDK_DIR)/go.mod.bak $(GO_SDK_DIR)/go.mod
	@echo "Go SDK generated in $(GO_SDK_DIR)"

# Test Python SDK build (produces .whl). Uses an isolated venv to avoid
# PEP 668 "externally-managed-environment" errors on system Python.
test_python_sdk: generate_python
	@echo "Testing Python SDK build..."
	cd $(PYTHON_SDK_DIR) && python -m venv .venv
	cd $(PYTHON_SDK_DIR) && .venv/bin/pip install build wheel
	cd $(PYTHON_SDK_DIR) && .venv/bin/python -m build
	@test -f $(PYTHON_SDK_DIR)/dist/*.whl || (echo "Python wheel not found" && exit 1)
	@rm -rf $(PYTHON_SDK_DIR)/.venv
	@echo "Python SDK build verified"

# Test Go SDK compilation.
test_go_sdk: generate_go
	@echo "Testing Go SDK build..."
	cd $(GO_SDK_DIR) && go mod tidy
	cd $(GO_SDK_DIR) && go build ./...
	@echo "Go SDK build verified"

# Unit tests for the bridge ProviderInfo (plan 05).
test:
	cd $(PROVIDER_DIR) && go test ./... -count=1

clean:
	rm -f $(TFGEN_BIN) $(RUNTIME_BIN)
	rm -f $(SCHEMA_DIR)/schema.json $(SCHEMA_DIR)/bridge-metadata.json $(SCHEMA_DIR)/schema-embed.json
