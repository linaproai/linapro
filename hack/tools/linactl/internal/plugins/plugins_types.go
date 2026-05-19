// This file defines shared plugin workspace types and runtime dependencies used
// by the official-plugin build path and configured source-plugin management.
// Package-level helpers keep the root command files free of plugin-specific
// structs while still exposing narrow interfaces for command orchestration.

package plugins

import (
	"context"
	"io"

	"linactl/internal/config"
	"linactl/internal/toolrun"
)

// Input exposes the command parameters needed by plugin workspace flows.
type Input interface {
	Get(string) string
	Has(string) bool
	Bool(string, bool) (bool, error)
}

// Runtime stores repository paths, process environment, streams, and command
// execution callbacks used by plugin management operations.
type Runtime struct {
	Root             string
	Env              []string
	Stdout           io.Writer
	Stderr           io.Writer
	RunCommand       toolrun.Runner
	RunCommandOutput toolrun.OutputRunner
}

// WorkspaceState classifies the official plugin workspace shape.
type WorkspaceState string

const (
	// WorkspaceStateMissing means apps/lina-plugins is absent.
	WorkspaceStateMissing WorkspaceState = "missing"
	// WorkspaceStateEmpty means apps/lina-plugins has no plugin manifests.
	WorkspaceStateEmpty WorkspaceState = "empty"
	// WorkspaceStateInvalid means apps/lina-plugins is present but unusable.
	WorkspaceStateInvalid WorkspaceState = "invalid"
	// WorkspaceStateReady means plugin manifests are discoverable.
	WorkspaceStateReady WorkspaceState = "ready"
)

// OfficialWorkspace describes the official plugin submodule checkout.
type OfficialWorkspace struct {
	Root          string
	State         WorkspaceState
	ManifestCount int
}

// Manifest stores the plugin fields needed by linactl.
type Manifest struct {
	Type    string `yaml:"type"`
	Version string `yaml:"version"`
}

// LoadConfig loads repository tool defaults for plugin management commands.
func LoadConfig(root string, input Input) (config.Root, error) {
	return config.Load(root, configPath(input))
}

// configPath resolves the optional config override used by plugin commands.
func configPath(input Input) string {
	if input == nil {
		return ""
	}
	return input.Get("config")
}

// run executes a child command through the configured runtime.
func (r Runtime) run(ctx context.Context, options toolrun.Options, name string, args ...string) error {
	return r.RunCommand(ctx, options, name, args...)
}

// output executes a child command through the configured runtime and returns stdout.
func (r Runtime) output(ctx context.Context, options toolrun.Options, name string, args ...string) (string, error) {
	return r.RunCommandOutput(ctx, options, name, args...)
}
