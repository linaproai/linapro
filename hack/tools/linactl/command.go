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
		{Name: "help", Description: "Show available cross-platform commands.", Usage: "linactl help [command]", Run: runHelp},
		{Name: "dev", Description: "Restart backend and frontend development services.", Usage: "linactl dev [backend_port=8080] [frontend_port=5666] [plugins=auto|0|1] [skip_wasm=true]", Run: runDev},
		{Name: "stop", Description: "Stop backend and frontend development services started by linactl.", Usage: "linactl stop [backend_port=8080] [frontend_port=5666]", Run: runStop},
		{Name: "status", Description: "Show backend and frontend service status.", Usage: "linactl status [backend_port=8080] [frontend_port=5666]", Run: runStatus},
		{Name: "prepare-packed-assets", Description: "Prepare host manifest assets for embedding.", Usage: "linactl prepare-packed-assets", Run: runPreparePackedAssets},
		{Name: "wasm", Description: "Build dynamic Wasm plugin artifacts.", Usage: "linactl wasm [p=<plugin-id>] [out=temp/output] [dry_run=true]", Run: runWasm},
		{Name: "build", Description: "Build frontend assets, plugin artifacts, and host binaries.", Usage: "linactl build [plugins=auto|0|1] [platforms=linux/amd64] [verbose=1]", Run: runBuild},
		{Name: "image", Description: "Build the production Docker image using existing image-builder.", Usage: "linactl image [tag=v0.6.0] [push=1]", Run: runImage},
		{Name: "image-build", Description: "Stage image build artifacts without invoking Docker build.", Usage: "linactl image-build [tag=v0.6.0]", Run: runImageBuild},
		{Name: "init", Description: "Initialize the database with DDL and seed data.", Usage: "linactl init confirm=init [rebuild=true]", Run: runInit},
		{Name: "mock", Description: "Load optional mock demo data.", Usage: "linactl mock confirm=mock", Run: runMock},
		{Name: "test", Description: "Run the Playwright E2E test suite.", Usage: "linactl test [scope=full|host|plugins|plugin:<id>]", Run: runTest},
		{Name: "test-go", Description: "Run Go unit tests for workspace modules.", Usage: "linactl test-go [plugins=auto|0|1] [race=true] [verbose=true]", Run: runTestGo},
		{Name: "tidy", Description: "Run go mod tidy in every maintained Go module directory.", Usage: "linactl tidy", Run: runTidy},
		{Name: "test-scripts", Description: "Run repository tool smoke tests.", Usage: "linactl test-scripts", Run: runTestScripts},
		{Name: "check-runtime-i18n", Description: "Scan runtime-visible code for hard-coded text.", Usage: "linactl check-runtime-i18n", Run: runCheckRuntimeI18n},
		{Name: "check-runtime-i18n-messages", Description: "Validate runtime i18n message key coverage.", Usage: "linactl check-runtime-i18n-messages", Run: runCheckRuntimeI18nMessages},
		{Name: "cli", Description: "Install or update the GoFrame CLI.", Usage: "linactl cli", Run: runCLIInstall},
		{Name: "cli.install", Description: "Install the GoFrame CLI only when missing.", Usage: "linactl cli.install", Run: runCLIInstallIfMissing},
		{Name: "ctrl", Description: "Generate GoFrame controllers.", Usage: "linactl ctrl", Run: runGF("gen", "ctrl")},
		{Name: "dao", Description: "Generate GoFrame DAO/DO/Entity files.", Usage: "linactl dao", Run: runGF("gen", "dao")},
		{Name: "enums", Description: "Generate GoFrame enum files.", Usage: "linactl enums", Run: runGF("gen", "enums")},
		{Name: "service", Description: "Generate GoFrame service files.", Usage: "linactl service", Run: runGF("gen", "service")},
		{Name: "pb", Description: "Generate protobuf files.", Usage: "linactl pb", Run: runGF("gen", "pb")},
		{Name: "pbentity", Description: "Generate protobuf entity files.", Usage: "linactl pbentity", Run: runGF("gen", "pbentity")},
		{Name: "deploy", Description: "Apply kustomize manifests to the current kubectl context.", Usage: "linactl deploy [env=<overlay>] [tag=develop]", Run: runDeploy},
	}

	registry := make(map[string]commandSpec, len(specs))
	for _, spec := range specs {
		registry[spec.Name] = spec
	}
	return registry
}

// normalizeCommandName converts historical make target aliases to linactl command names.
func normalizeCommandName(name string) string {
	switch strings.TrimSpace(name) {
	case "prepare":
		return "prepare-packed-assets"
	default:
		return strings.TrimSpace(name)
	}
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
func runHelp(_ context.Context, a *app, _ commandInput) error {
	return a.printHelp()
}

// printHelp writes the command overview and platform-specific entry examples.
func (a *app) printHelp() error {
	specs := commandRegistry()
	names := make([]string, 0, len(specs))
	for name := range specs {
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
	fmt.Fprintln(a.stdout, "  cmd.exe:     make.cmd help")
	fmt.Fprintln(a.stdout, "  PowerShell:  .\\make.cmd help")
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
