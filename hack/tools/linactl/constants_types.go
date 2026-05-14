// This file defines shared linactl constants and data types.

package main

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"time"
)

const (
	// defaultBackendPort is the standard backend development port.
	defaultBackendPort = 8080
	// defaultFrontendPort is the standard frontend development port.
	defaultFrontendPort = 5666
	// defaultWaitTimeout bounds development service readiness checks.
	defaultWaitTimeout = 60 * time.Second
	// officialPluginsBuildTag enables compiled official source-plugin backends.
	officialPluginsBuildTag = "official_plugins"
	// sourcePluginsEnvKey controls frontend source-plugin page discovery.
	sourcePluginsEnvKey = "LINAPRO_SOURCE_PLUGINS"
	// officialPluginInitCommand is the operator-facing submodule bootstrap command.
	officialPluginInitCommand = "git submodule update --init --recursive"
	// officialPluginWorkspaceFile is the ignored temporary Go workspace for plugin-full builds.
	officialPluginWorkspaceFile = "go.work.plugins"
)

// errHelpRequested marks help output as a successful early return.
var errHelpRequested = errors.New("help requested")

// commandSpec describes one supported linactl command.
type commandSpec struct {
	Name        string
	Description string
	Usage       string
	Run         func(context.Context, *app, commandInput) error
}

// commandInput stores parsed command arguments.
type commandInput struct {
	Args   []string
	Params map[string]string
}

// app stores one linactl invocation's process dependencies and repository paths.
type app struct {
	stdout io.Writer
	stderr io.Writer
	stdin  io.Reader

	root string
	env  []string

	execCommand func(context.Context, string, ...string) *exec.Cmd
	waitHTTP    func(string, string, string, string, time.Duration) error
}

// rootConfig stores repository-level tool configuration from hack/config.yaml.
type rootConfig struct {
	Build buildConfig `yaml:"build"`
	Image imageConfig `yaml:"image"`
}

// buildConfig stores user-facing build defaults.
type buildConfig struct {
	Platforms  []string `yaml:"platforms"`
	CGOEnabled bool     `yaml:"cgoEnabled"`
	OutputDir  string   `yaml:"outputDir"`
	BinaryName string   `yaml:"binaryName"`
}

// imageConfig stores user-facing image metadata defaults.
type imageConfig struct {
	Name       string `yaml:"name"`
	Tag        string `yaml:"tag"`
	Registry   string `yaml:"registry"`
	Push       bool   `yaml:"push"`
	BaseImage  string `yaml:"baseImage"`
	Dockerfile string `yaml:"dockerfile"`
}

// targetPlatform stores one normalized Go target platform.
type targetPlatform struct {
	OS   string
	Arch string
}

// serviceConfig stores development service paths and ports.
type serviceConfig struct {
	Name      string
	URL       string
	Port      int
	PIDPath   string
	LogPath   string
	WorkDir   string
	StartName string
	StartArgs []string
}

// serviceStatusRow stores one printable development service status row.
type serviceStatusRow struct {
	Service string
	Status  string
	URL     string
	PID     string
	PIDFile string
	LogFile string
}

// pluginWorkspaceState classifies the official plugin workspace shape.
type pluginWorkspaceState string

const (
	// pluginWorkspaceStateMissing means apps/lina-plugins is absent.
	pluginWorkspaceStateMissing pluginWorkspaceState = "missing"
	// pluginWorkspaceStateEmpty means apps/lina-plugins has no plugin manifests.
	pluginWorkspaceStateEmpty pluginWorkspaceState = "empty"
	// pluginWorkspaceStateInvalid means apps/lina-plugins is present but unusable.
	pluginWorkspaceStateInvalid pluginWorkspaceState = "invalid"
	// pluginWorkspaceStateReady means plugin manifests are discoverable.
	pluginWorkspaceStateReady pluginWorkspaceState = "ready"
)

// officialPluginWorkspace describes the official plugin submodule checkout.
type officialPluginWorkspace struct {
	Root          string
	State         pluginWorkspaceState
	ManifestCount int
}

// pluginManifest stores the plugin fields needed by linactl.
type pluginManifest struct {
	Type string `yaml:"type"`
}
