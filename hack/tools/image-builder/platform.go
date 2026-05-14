// This file parses and formats Go and Docker target platforms.

package main

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// targetPlatform stores one normalized Go and Docker target platform.
type targetPlatform struct {
	OS   string
	Arch string
}

// splitPlatformCSV parses a command-line comma-separated platform override.
func splitPlatformCSV(value string) ([]string, error) {
	items := strings.Split(value, ",")
	platforms := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			return nil, errors.New("build.platforms contains an empty platform entry")
		}
		platforms = append(platforms, normalized)
	}
	if len(platforms) == 0 {
		return nil, errors.New("build.platforms cannot be empty")
	}
	return platforms, nil
}

// parsePlatformList parses one Docker/Go platform list from configuration.
func parsePlatformList(values []string) ([]targetPlatform, error) {
	if len(values) == 0 {
		return nil, errors.New("build.platforms cannot be empty")
	}
	targets := make([]targetPlatform, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		target, err := parseTargetPlatform(value)
		if err != nil {
			return nil, err
		}
		key := target.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		targets = append(targets, target)
	}
	if len(targets) == 0 {
		return nil, errors.New("build.platforms cannot be empty")
	}
	return targets, nil
}

// parseTargetPlatform parses one target platform in goos/goarch form.
func parseTargetPlatform(value string) (targetPlatform, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return targetPlatform{}, errors.New("build.platforms contains an empty platform entry")
	}
	if normalized == "auto" {
		return targetPlatform{OS: "linux", Arch: runtime.GOARCH}, nil
	}
	parts := strings.Split(normalized, "/")
	if len(parts) != 2 {
		return targetPlatform{}, fmt.Errorf("build.platforms entry %q must use goos/goarch form or auto", value)
	}
	target := targetPlatform{
		OS:   strings.TrimSpace(parts[0]),
		Arch: strings.TrimSpace(parts[1]),
	}
	if target.OS == "" || target.Arch == "" {
		return targetPlatform{}, fmt.Errorf("build.platforms entry %q must include both goos and goarch", value)
	}
	if strings.ContainsAny(target.OS+target.Arch, " \t\r\n") {
		return targetPlatform{}, fmt.Errorf("build.platforms entry %q must not contain whitespace", value)
	}
	return target, nil
}

// String returns the canonical goos/goarch representation.
func (p targetPlatform) String() string {
	return p.OS + "/" + p.Arch
}

// DirName returns the filesystem-safe directory segment for build outputs.
func (p targetPlatform) DirName() string {
	return p.OS + "_" + p.Arch
}

// joinPlatformCSV joins targets in Docker platform-list format.
func joinPlatformCSV(targets []targetPlatform) string {
	values := make([]string, 0, len(targets))
	for _, target := range targets {
		values = append(values, target.String())
	}
	return strings.Join(values, ",")
}

// platformValues returns normalized platform strings for the in-memory config.
func platformValues(targets []targetPlatform) []string {
	values := make([]string, 0, len(targets))
	for _, target := range targets {
		values = append(values, target.String())
	}
	return values
}

// joinPlatformSpace joins targets for shell for-loops in make recipes.
func joinPlatformSpace(targets []targetPlatform) string {
	values := make([]string, 0, len(targets))
	for _, target := range targets {
		values = append(values, target.String())
	}
	return strings.Join(values, " ")
}
