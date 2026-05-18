// This file manages environment variables for host-only and plugin-full Go builds.

package main

import "strings"

// officialPluginBuildEnv returns the process environment for host-only or
// plugin-full frontend and backend builds.
func officialPluginBuildEnv(root string, env []string, enabled bool, workspacePath string) []string {
	value := "0"
	env = removeOfficialPluginBuildTag(env)
	if enabled {
		value = "1"
		if workspacePath == "" {
			workspacePath = officialPluginGoWorkPath(root)
		}
		env = removeEnvValue(env, "GOWORK")
		env = setEnvValue(env, "GOWORK", workspacePath)
		env = appendGoBuildTag(env, officialPluginsBuildTag)
	} else {
		env = removeEnvValue(env, "GOWORK")
	}
	return setEnvValue(env, sourcePluginsEnvKey, value)
}

// removeOfficialPluginBuildTag removes the official plugin build tag so
// host-only commands are not affected by a developer's inherited GOFLAGS.
func removeOfficialPluginBuildTag(env []string) []string {
	current := strings.TrimSpace(envValue(env, "GOFLAGS"))
	if current == "" {
		return env
	}
	parts := strings.Fields(current)
	next := make([]string, 0, len(parts))
	for index := 0; index < len(parts); index++ {
		part := parts[index]
		if strings.HasPrefix(part, "-tags=") {
			tags := removeGoBuildTagValue(strings.TrimPrefix(part, "-tags="), officialPluginsBuildTag)
			if tags != "" {
				next = append(next, "-tags="+tags)
			}
			continue
		}
		if part == "-tags" && index+1 < len(parts) {
			index++
			tags := removeGoBuildTagValue(parts[index], officialPluginsBuildTag)
			if tags != "" {
				next = append(next, "-tags", tags)
			}
			continue
		}
		next = append(next, part)
	}
	if len(next) == 0 {
		return removeEnvValue(env, "GOFLAGS")
	}
	return setEnvValue(env, "GOFLAGS", strings.Join(next, " "))
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
	current := strings.TrimSpace(envValue(env, "GOFLAGS"))
	if current == "" {
		return setEnvValue(env, "GOFLAGS", "-tags="+tag)
	}
	parts := strings.Fields(current)
	for index := 0; index < len(parts); index++ {
		part := parts[index]
		if strings.HasPrefix(part, "-tags=") {
			parts[index] = "-tags=" + addGoBuildTagValue(strings.TrimPrefix(part, "-tags="), tag)
			return setEnvValue(env, "GOFLAGS", strings.Join(parts, " "))
		}
		if part == "-tags" && index+1 < len(parts) {
			parts[index+1] = addGoBuildTagValue(parts[index+1], tag)
			return setEnvValue(env, "GOFLAGS", strings.Join(parts, " "))
		}
	}
	return setEnvValue(env, "GOFLAGS", strings.TrimSpace(current+" -tags="+tag))
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

// envValue returns one environment value from a KEY=VALUE list.
func envValue(env []string, key string) string {
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return strings.TrimPrefix(item, prefix)
		}
	}
	return ""
}

// setEnvValue returns a copy of env with one key replaced or appended.
func setEnvValue(env []string, key string, value string) []string {
	prefix := key + "="
	next := append([]string{}, env...)
	for index, item := range next {
		if strings.HasPrefix(item, prefix) {
			next[index] = prefix + value
			return next
		}
	}
	return append(next, prefix+value)
}

// removeEnvValue returns a copy of env without one key.
func removeEnvValue(env []string, key string) []string {
	prefix := key + "="
	next := make([]string, 0, len(env))
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			continue
		}
		next = append(next, item)
	}
	return next
}
