// This file defines command registration, help output, and argument parsing.

package main

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// commandRegistry returns the supported command list keyed by command name.
func commandRegistry() map[string]commandSpec {
	specs := []commandSpec{
		{Name: "help", Description: "Show available cross-platform commands.", Usage: "linactl help [command|--all]", Run: runHelp},
		{Name: "dev", Description: "Restart backend and frontend development services.", Usage: "linactl dev [backend_port=8080] [frontend_port=5666] [plugins=auto|0|1] [skip_wasm=true]", Run: runDev},
		{Name: "stop", Description: "Stop backend and frontend development services started by linactl.", Usage: "linactl stop [backend_port=8080] [frontend_port=5666]", Run: runStop},
		{Name: "status", Description: "Show backend and frontend service status.", Usage: "linactl status [backend_port=8080] [frontend_port=5666]", Run: runStatus},
		{Name: "pack.assets", Description: "Prepare host manifest assets for embedding.", Usage: "linactl pack.assets", Run: runPreparePackedAssets},
		{Name: "wasm", Description: "Build dynamic Wasm plugin artifacts.", Usage: "linactl wasm [p=<plugin-id>] [out=temp/output] [dry_run=true]", Run: runWasm},
		{Name: "plugins.init", Description: "Convert apps/lina-plugins from a submodule to a normal plugin directory.", Usage: "linactl plugins.init", Run: runPluginsInit},
		{Name: "plugins.install", Description: "Install configured source plugins into apps/lina-plugins.", Usage: "linactl plugins.install [p=<plugin-id>] [source=<name>] [force=1]", Run: runPluginsInstall},
		{Name: "plugins.update", Description: "Update configured source plugins in apps/lina-plugins.", Usage: "linactl plugins.update [p=<plugin-id>] [source=<name>] [force=1]", Run: runPluginsUpdate},
		{Name: "plugins.status", Description: "Show configured source-plugin workspace status.", Usage: "linactl plugins.status [p=<plugin-id>] [source=<name>]", Run: runPluginsStatus},
		{Name: "build", Description: "Build frontend assets, plugin artifacts, and host binaries.", Usage: "linactl build [plugins=auto|0|1] [platforms=linux/amd64] [verbose=1]", Run: runBuild},
		{Name: "image", Description: "Build the production Docker image using existing image-builder.", Usage: "linactl image [tag=v0.6.0] [push=1]", Run: runImage},
		{Name: "image.build", Description: "Stage image build artifacts without invoking Docker build.", Usage: "linactl image.build [tag=v0.6.0]", Run: runImageBuild},
		{Name: "release.tag.check", Description: "Verify a release tag matches framework.version metadata.", Usage: "linactl release.tag.check [tag=v0.6.0]", Run: runReleaseTagCheck},
		{Name: "init", Description: "Initialize the database with DDL and seed data.", Usage: "linactl init confirm=init [rebuild=true]", Run: runInit},
		{Name: "mock", Description: "Load optional mock demo data.", Usage: "linactl mock confirm=mock", Run: runMock},
		{Name: "test", Description: "Run the Playwright E2E test suite.", Usage: "linactl test [scope=full|host|plugins|plugin:<id>]", Run: runTest},
		{Name: "test.go", Description: "Run Go unit tests for workspace modules.", Usage: "linactl test.go [plugins=auto|0|1] [race=true] [verbose=true]", Run: runTestGo},
		{Name: "test.host", Description: "Run host-owned Playwright E2E tests without official plugins.", Usage: "linactl test.host", Run: runTestHost},
		{Name: "test.plugins", Description: "Run official plugin Playwright E2E tests.", Usage: "linactl test.plugins", Run: runTestPlugins},
		{Name: "tidy", Description: "Run go mod tidy in every maintained Go module directory.", Usage: "linactl tidy", Run: runTidy},
		{Name: "test.scripts", Description: "Run repository tool smoke tests.", Usage: "linactl test.scripts", Run: runTestScripts},
		{Name: "i18n.check", Description: "Run runtime i18n hard-coded text and message coverage checks.", Usage: "linactl i18n.check", Run: runI18nCheck},
		{Name: "cli", Description: "Install or update the GoFrame CLI.", Usage: "linactl cli", Internal: true, Run: runCLIInstall},
		{Name: "cli.install", Description: "Install the GoFrame CLI only when missing.", Usage: "linactl cli.install", Internal: true, Run: runCLIInstallIfMissing},
		{Name: "ctrl", Description: "Generate GoFrame controllers.", Usage: "linactl ctrl", Internal: true, Run: runGF("gen", "ctrl")},
		{Name: "dao", Description: "Generate GoFrame DAO/DO/Entity files.", Usage: "linactl dao", Internal: true, Run: runGF("gen", "dao")},
	}

	registry := make(map[string]commandSpec, len(specs))
	for _, spec := range specs {
		registry[spec.Name] = spec
	}
	return registry
}

// normalizeCommandName canonicalizes command names before registry lookup.
func normalizeCommandName(name string) string {
	return strings.TrimSpace(name)
}

// parseCommandInput accepts make-style key=value parameters and standard flags.
func parseCommandInput(args []string) (commandInput, error) {
	input := commandInput{Params: map[string]string{}}
	for _, arg := range args {
		if arg == "" {
			continue
		}
		if strings.HasPrefix(arg, "--") {
			trimmed := strings.TrimPrefix(arg, "--")
			if trimmed == "" {
				return input, fmt.Errorf("invalid empty flag")
			}
			key, value, ok := strings.Cut(trimmed, "=")
			key = normalizeParamKey(key)
			if !ok {
				input.Params[key] = "true"
				continue
			}
			if key == "" {
				return input, fmt.Errorf("invalid flag %q", arg)
			}
			input.Params[key] = value
			continue
		}
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			input.Params[normalizeParamKey(strings.TrimPrefix(arg, "-"))] = "true"
			continue
		}
		if key, value, ok := strings.Cut(arg, "="); ok {
			key = normalizeParamKey(key)
			if key == "" {
				return input, fmt.Errorf("invalid parameter %q", arg)
			}
			input.Params[key] = value
			continue
		}
		input.Args = append(input.Args, arg)
	}
	return input, nil
}

// Get returns a parsed parameter value.
func (i commandInput) Get(key string) string {
	return i.Params[normalizeParamKey(key)]
}

// Has reports whether a parameter was explicitly provided.
func (i commandInput) Has(key string) bool {
	_, ok := i.Params[normalizeParamKey(key)]
	return ok
}

// GetDefault returns a parameter value or the provided default.
func (i commandInput) GetDefault(key string, fallback string) string {
	if value, ok := i.Params[normalizeParamKey(key)]; ok && value != "" {
		return value
	}
	return fallback
}

// HasBool reports whether a flag-style boolean parameter is true.
func (i commandInput) HasBool(key string) bool {
	value, ok := i.Params[normalizeParamKey(key)]
	if !ok {
		return false
	}
	parsed, err := parseBool(value, false)
	if err != nil {
		return false
	}
	return parsed
}

// Bool returns a parsed boolean parameter.
func (i commandInput) Bool(key string, fallback bool) (bool, error) {
	value, ok := i.Params[normalizeParamKey(key)]
	if !ok {
		return fallback, nil
	}
	return parseBool(value, fallback)
}

// runHelp prints the top-level help output.
func runHelp(_ context.Context, a *app, input commandInput) error {
	return a.printHelp(input.HasBool("all"))
}

// printHelp writes the command overview and platform-specific entry examples.
func (a *app) printHelp(includeInternal bool) error {
	specs := commandRegistry()
	names := make([]string, 0, len(specs))
	for name, spec := range specs {
		if spec.Internal && !includeInternal {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	maxName := 0
	for _, name := range names {
		if len(name) > maxName {
			maxName = len(name)
		}
	}

	fmt.Fprintln(a.stdout, "Usage: linactl <command> [key=value] [--flag=value]")
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Windows:")
	fmt.Fprintln(a.stdout, "  cmd.exe:     make help")
	fmt.Fprintln(a.stdout, "  PowerShell:  .\\make help")
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Linux/macOS:")
	fmt.Fprintln(a.stdout, "  make help")
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Available commands:")
	for _, name := range names {
		spec := specs[name]
		fmt.Fprintf(a.stdout, "  %-*s  %s\n", maxName, spec.Name, spec.Description)
	}
	return nil
}

// printCommandHelp writes usage for one command.
func printCommandHelp(out io.Writer, spec commandSpec) {
	fmt.Fprintf(out, "Usage: %s\n\n%s\n", spec.Usage, spec.Description)
}

// Int returns a parsed integer parameter.
func (i commandInput) Int(key string, fallback int) (int, error) {
	value, ok := i.Params[normalizeParamKey(key)]
	if !ok || value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}
