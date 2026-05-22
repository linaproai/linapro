// This file defines command registration and argument parsing for linactl.

package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"linactl/internal/toolutil"
)

// commandRegistry returns the supported command list keyed by command name.
func commandRegistry() map[string]commandSpec {
	specs := []commandSpec{
		{Name: "help", Description: "Show available cross-platform commands.", Usage: "linactl help [command|--all]", Run: runHelp},
		{Name: "dev", Description: "Restart backend and frontend development services.", Usage: "linactl dev [backend_port=9120] [frontend_port=5666] [plugins=auto|0|1] [skip_wasm=true]", Run: runDev},
		{Name: "stop", Description: "Stop backend and frontend development services started by linactl.", Usage: "linactl stop [backend_port=9120] [frontend_port=5666]", Run: runStop},
		{Name: "status", Description: "Show backend and frontend service status.", Usage: "linactl status [backend_port=9120] [frontend_port=5666]", Run: runStatus},
		{Name: "pack.assets", Description: "Prepare host manifest assets for embedding.", Usage: "linactl pack.assets", Run: runPreparePackedAssets},
		{Name: "wasm", Description: "Build dynamic Wasm plugin artifacts.", Usage: "linactl wasm [p=<plugin-id>|plugin_dir=<path>] [out=temp/output] [dry_run=true]", Run: runWasm},
		{Name: "plugins.init", Description: "Convert apps/lina-plugins from a submodule to a normal plugin directory.", Usage: "linactl plugins.init", Run: runPluginsInit},
		{Name: "plugins.install", Description: "Install configured source plugins into apps/lina-plugins.", Usage: "linactl plugins.install [p=<plugin-id>] [source=<name>] [force=1]", Run: runPluginsInstall},
		{Name: "plugins.update", Description: "Update configured source plugins in apps/lina-plugins.", Usage: "linactl plugins.update [p=<plugin-id>] [source=<name>] [force=1]", Run: runPluginsUpdate},
		{Name: "plugins.status", Description: "Show configured source-plugin workspace status.", Usage: "linactl plugins.status [p=<plugin-id>] [source=<name>]", Run: runPluginsStatus},
		{Name: "build", Description: "Build frontend assets, plugin artifacts, and host binaries.", Usage: "linactl build [plugins=auto|0|1] [platforms=linux/amd64] [verbose=1]", Run: runBuild},
		{Name: "image", Description: "Build the production Docker image.", Usage: "linactl image [tag=v0.6.0] [push=1]", Run: runImage},
		{Name: "image.build", Description: "Stage image build artifacts without invoking Docker build.", Usage: "linactl image.build [tag=v0.6.0]", Run: runImageBuild},
		{Name: "release.tag.check", Description: "Verify a release tag matches framework.version metadata.", Usage: "linactl release.tag.check [tag=v0.6.0]", Run: runReleaseTagCheck},
		{Name: "env.check", Description: "Check local development tool versions.", Usage: "linactl env.check", Run: runEnvCheck},
		{Name: "env.setup", Description: "Install frontend dependencies and Playwright browsers.", Usage: "linactl env.setup", Run: runEnvSetup},
		{Name: "init", Description: "Initialize the database with DDL and seed data.", Usage: "linactl init confirm=init [rebuild=true]", Run: runInit},
		{Name: "mock", Description: "Load optional mock demo data.", Usage: "linactl mock confirm=mock", Run: runMock},
		{Name: "test", Description: "Run the Playwright E2E test suite.", Usage: "linactl test [scope=full|host|plugins|plugin:<id>]", Run: runTest},
		{Name: "test.go", Description: "Run Go unit tests for workspace modules.", Usage: "linactl test.go [plugins=auto|0|1] [race=true] [verbose=true]", Run: runTestGo},
		{Name: "test.host", Description: "Run host-owned Playwright E2E tests without official plugins.", Usage: "linactl test.host", Run: runTestHost},
		{Name: "test.plugins", Description: "Run official plugin Playwright E2E tests.", Usage: "linactl test.plugins", Run: runTestPlugins},
		{Name: "tidy", Description: "Run go mod tidy in every maintained Go module directory.", Usage: "linactl tidy", Run: runTidy},
		{Name: "test.scripts", Description: "Run repository tool smoke tests.", Usage: "linactl test.scripts", Run: runTestScripts},
		{Name: "i18n.check", Description: "Run runtime i18n hard-coded text and message coverage checks.", Usage: "linactl i18n.check", Run: runI18nCheck},
		{Name: "agents", Description: "One-shot agent setup: arrow-key picker on TTY, or agent=<name> for non-interactive use.", Usage: "linactl agents [agent=<name>] [action=link|unlink] [force=1]", Run: runAgents},
		{Name: "agents.skills.link", Description: "Manage repository-local symlinks from supported agents' project skill paths to .agents/skills.", Usage: "linactl agents.skills.link [agent=<name|all|csv>] [force=1]", Run: runAgentsSkillsLink},
		{Name: "agents.skills.unlink", Description: "Remove repository-local skills symlinks managed by agents.skills.link.", Usage: "linactl agents.skills.unlink agent=<name|all|csv>", Run: runAgentsSkillsUnlink},
		{Name: "agents.prompts.link", Description: "Manage repository-local symlinks from supported agents' commands/prompts roots to .agents/prompts.", Usage: "linactl agents.prompts.link [agent=<name|all|csv>] [force=1]", Run: runAgentsPromptsLink},
		{Name: "agents.prompts.unlink", Description: "Remove repository-local prompts symlinks managed by agents.prompts.link.", Usage: "linactl agents.prompts.unlink agent=<name|all|csv>", Run: runAgentsPromptsUnlink},
		{Name: "agents.md.link", Description: "Manage repository-local symlinks from supported agents' private guide files to AGENTS.md.", Usage: "linactl agents.md.link [agent=<name|all|csv>] [force=1]", Run: runAgentsMdLink},
		{Name: "agents.md.unlink", Description: "Remove repository-local AGENTS.md symlinks managed by agents.md.link.", Usage: "linactl agents.md.unlink agent=<name|all|csv>", Run: runAgentsMdUnlink},
		{Name: "cli", Description: "Install or update the GoFrame CLI.", Usage: "linactl cli", Internal: true, Run: runCLIInstall},
		{Name: "cli.install", Description: "Install the GoFrame CLI only when missing.", Usage: "linactl cli.install", Internal: true, Run: runCLIInstallIfMissing},
		{Name: "ctrl", Description: "Generate GoFrame controllers.", Usage: "linactl ctrl", Internal: true, Run: runCtrl},
		{Name: "dao", Description: "Generate GoFrame DAO/DO/Entity files.", Usage: "linactl dao", Internal: true, Run: runDao},
	}

	registry := make(map[string]commandSpec, len(specs))
	for _, spec := range specs {
		registry[spec.Name] = spec
	}
	return registry
}

// commandNames returns the registered command names in deterministic order for
// governance scans that need to verify command implementation files.
func commandNames() []string {
	registry := commandRegistry()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
			key = toolutil.NormalizeParamKey(key)
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
			input.Params[toolutil.NormalizeParamKey(strings.TrimPrefix(arg, "-"))] = "true"
			continue
		}
		if key, value, ok := strings.Cut(arg, "="); ok {
			key = toolutil.NormalizeParamKey(key)
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
	return i.Params[toolutil.NormalizeParamKey(key)]
}

// Has reports whether a parameter was explicitly provided.
func (i commandInput) Has(key string) bool {
	_, ok := i.Params[toolutil.NormalizeParamKey(key)]
	return ok
}

// GetDefault returns a parameter value or the provided default.
func (i commandInput) GetDefault(key string, fallback string) string {
	if value, ok := i.Params[toolutil.NormalizeParamKey(key)]; ok && value != "" {
		return value
	}
	return fallback
}

// HasBool reports whether a flag-style boolean parameter is true.
func (i commandInput) HasBool(key string) bool {
	value, ok := i.Params[toolutil.NormalizeParamKey(key)]
	if !ok {
		return false
	}
	parsed, err := toolutil.ParseBool(value, false)
	if err != nil {
		return false
	}
	return parsed
}

// Bool returns a parsed boolean parameter.
func (i commandInput) Bool(key string, fallback bool) (bool, error) {
	value, ok := i.Params[toolutil.NormalizeParamKey(key)]
	if !ok {
		return fallback, nil
	}
	return toolutil.ParseBool(value, fallback)
}

// Int returns a parsed integer parameter.
func (i commandInput) Int(key string, fallback int) (int, error) {
	value, ok := i.Params[toolutil.NormalizeParamKey(key)]
	if !ok || value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

// ParamMap returns the normalized parameter map for internal components that
// need to iterate over all values.
func (i commandInput) ParamMap() map[string]string {
	return i.Params
}
