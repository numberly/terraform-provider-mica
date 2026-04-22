PROJECT_NAME := pulumi-flashblade
PROVIDER_PKG := github.com/numberly/opentofu-provider-flashblade/pulumi/provider
VERSION := $(shell git describe --tags --dirty --always 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X $(PROVIDER_PKG)/pkg/version.Version=$(VERSION)"

PROVIDER_DIR := provider
TFGEN_BIN   := $(PROVIDER_DIR)/cmd/pulumi-tfgen-flashblade/pulumi-tfgen-flashblade
RUNTIME_BIN := $(PROVIDER_DIR)/cmd/pulumi-resource-flashblade/pulumi-resource-flashblade
SCHEMA_DIR  := $(PROVIDER_DIR)/cmd/pulumi-resource-flashblade

.PHONY: version tfgen_build tfgen provider test clean

version:
	@echo $(VERSION)

# Build tfgen binary (no ldflags version injection — tfgen is a build tool).
tfgen_build:
	cd $(PROVIDER_DIR) && go build -o cmd/pulumi-tfgen-flashblade/pulumi-tfgen-flashblade ./cmd/pulumi-tfgen-flashblade

# Generate schema artifacts. tfgen's `schema` subcommand emits schema.json + bridge-metadata.json
# + schema-embed.json into the --out directory.
tfgen: tfgen_build
	$(TFGEN_BIN) schema --out $(SCHEMA_DIR)
	@test -s $(SCHEMA_DIR)/schema.json        || (echo "schema.json empty" && exit 1)
	@test -s $(SCHEMA_DIR)/schema-embed.json  || (echo "schema-embed.json empty" && exit 1)
	@test -s $(SCHEMA_DIR)/bridge-metadata.json || (echo "bridge-metadata.json empty" && exit 1)

# Build runtime provider binary with ldflags-injected VERSION and real embedded schema.
provider: tfgen
	cd $(PROVIDER_DIR) && go build $(LDFLAGS) -o cmd/pulumi-resource-flashblade/pulumi-resource-flashblade ./cmd/pulumi-resource-flashblade

# Unit tests for the bridge ProviderInfo (plan 05).
test:
	cd $(PROVIDER_DIR) && go test ./... -count=1

clean:
	rm -f $(TFGEN_BIN) $(RUNTIME_BIN)
	rm -f $(SCHEMA_DIR)/schema.json $(SCHEMA_DIR)/schema-embed.json $(SCHEMA_DIR)/bridge-metadata.json
