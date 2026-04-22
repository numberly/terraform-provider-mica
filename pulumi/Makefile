PROJECT_NAME := pulumi-flashblade
PROVIDER_PKG := github.com/numberly/opentofu-provider-flashblade/pulumi/provider
VERSION := $(shell git describe --tags --dirty --always 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X $(PROVIDER_PKG)/pkg/version.Version=$(VERSION)"

PROVIDER_DIR := provider
TFGEN_BIN   := $(PROVIDER_DIR)/cmd/pulumi-tfgen-flashblade/pulumi-tfgen-flashblade
RUNTIME_BIN := $(PROVIDER_DIR)/cmd/pulumi-resource-flashblade/pulumi-resource-flashblade
SCHEMA_DIR  := $(PROVIDER_DIR)/cmd/pulumi-resource-flashblade

.PHONY: version tfgen provider test clean

version:
	@echo $(VERSION)

# tfgen, provider, test, clean implementations are added by plan 04 + plan 05.
tfgen:
	@echo "tfgen target not yet implemented — see plan 04"
	@exit 1

provider:
	@echo "provider target not yet implemented — see plan 04"
	@exit 1

test:
	@echo "test target not yet implemented — see plan 05"
	@exit 1

clean:
	rm -f $(TFGEN_BIN) $(RUNTIME_BIN)
	rm -f $(SCHEMA_DIR)/schema.json $(SCHEMA_DIR)/schema-embed.json $(SCHEMA_DIR)/bridge-metadata.json
