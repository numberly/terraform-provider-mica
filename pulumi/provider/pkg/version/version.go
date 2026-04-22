// Package version exposes the provider build version.
// Value is injected at build time via -ldflags "-X .../pkg/version.Version=<git describe>".
package version

var Version = "0.0.1"
