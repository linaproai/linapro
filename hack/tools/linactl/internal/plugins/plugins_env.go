// Package plugins manages official and user-project source plugin workspace
// flows for linactl. This file owns environment mutation for host-only and
// plugin-full builds, keeping build-tag handling away from command files.
package plugins

import (
	"strings"

	"linactl/internal/toolutil"
)

const (
	// OfficialBuildTag enables compiled official source-plugin backends.
	OfficialBuildTag = "official_plugins"
	// SourcePluginsEnvKey controls frontend source-plugin page discovery.
	SourcePluginsEnvKey = "LINAPRO_SOURCE_PLUGINS"
	// InitCommand is the operator-facing submodule bootstrap command.
	InitCommand = "git submodule update --init --recursive"
	// WorkspaceFile is the ignored temporary Go workspace for plugin-full builds.
	WorkspaceFile = "go.work.plugins"
	// AggregateModuleName is the import path used by the host plugin registry bridge.
	AggregateModuleName = "lina-plugins"
	// AggregateDir is the ignored generated module that imports source-plugin backends.
	AggregateDir = "official-plugins"
)

// BuildEnv returns the process environment for host-only or
// plugin-full frontend and backend builds.
func BuildEnv(root string, env []string, enabled bool, workspacePath string) []string {
	value := "0"
	env = removeOfficialPluginBuildTag(env)
	if enabled {
		value = "1"
		if workspacePath == "" {
			workspacePath = GoWorkPath(root)
		}
		env = toolutil.RemoveEnvValue(env, "GOWORK")
		env = toolutil.SetEnvValue(env, "GOWORK", workspacePath)
		env = appendGoBuildTag(env, OfficialBuildTag)
	} else {
		env = toolutil.RemoveEnvValue(env, "GOWORK")
	}
	return toolutil.SetEnvValue(env, SourcePluginsEnvKey, value)
}

// removeOfficialPluginBuildTag removes the official plugin build tag so
// host-only commands are not affected by a developer's inherited GOFLAGS.
func removeOfficialPluginBuildTag(env []string) []string {
	current := strings.TrimSpace(toolutil.EnvValue(env, "GOFLAGS"))
	if current == "" {
		return env
	}
	parts := strings.Fields(current)
	next := make([]string, 0, len(parts))
	for index := 0; index < len(parts); index++ {
		part := parts[index]
		if strings.HasPrefix(part, "-tags=") {
			tags := removeGoBuildTagValue(strings.TrimPrefix(part, "-tags="), OfficialBuildTag)
			if tags != "" {
				next = append(next, "-tags="+tags)
			}
			continue
		}
		if part == "-tags" && index+1 < len(parts) {
			index++
			tags := removeGoBuildTagValue(parts[index], OfficialBuildTag)
			if tags != "" {
				next = append(next, "-tags", tags)
			}
			continue
		}
		next = append(next, part)
	}
	if len(next) == 0 {
		return toolutil.RemoveEnvValue(env, "GOFLAGS")
	}
	return toolutil.SetEnvValue(env, "GOFLAGS", strings.Join(next, " "))
}

// removeGoBuildTagValue removes one comma-separated Go build tag from a flag value.
func removeGoBuildTagValue(value string, tag string) string {
	items := strings.Split(value, ",")
	next := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" || trimmed == tag {
			continue
		}
		next = append(next, trimmed)
	}
	return strings.Join(next, ",")
}

// appendGoBuildTag appends one build tag to GOFLAGS without discarding existing flags.
func appendGoBuildTag(env []string, tag string) []string {
	current := strings.TrimSpace(toolutil.EnvValue(env, "GOFLAGS"))
	if current == "" {
		return toolutil.SetEnvValue(env, "GOFLAGS", "-tags="+tag)
	}
	parts := strings.Fields(current)
	for index := 0; index < len(parts); index++ {
		part := parts[index]
		if strings.HasPrefix(part, "-tags=") {
			parts[index] = "-tags=" + addGoBuildTagValue(strings.TrimPrefix(part, "-tags="), tag)
			return toolutil.SetEnvValue(env, "GOFLAGS", strings.Join(parts, " "))
		}
		if part == "-tags" && index+1 < len(parts) {
			parts[index+1] = addGoBuildTagValue(parts[index+1], tag)
			return toolutil.SetEnvValue(env, "GOFLAGS", strings.Join(parts, " "))
		}
	}
	return toolutil.SetEnvValue(env, "GOFLAGS", strings.TrimSpace(current+" -tags="+tag))
}

// addGoBuildTagValue appends one comma-separated Go build tag if missing.
func addGoBuildTagValue(value string, tag string) string {
	items := strings.Split(value, ",")
	next := make([]string, 0, len(items)+1)
	found := false
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if trimmed == tag {
			found = true
		}
		next = append(next, trimmed)
	}
	if !found {
		next = append(next, tag)
	}
	return strings.Join(next, ",")
}
