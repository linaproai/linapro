// This file implements release governance commands for repository metadata.

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// releaseVersionPattern accepts the Docker-compatible release version subset.
var releaseVersionPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z][0-9A-Za-z.-]*)?$`)

// dockerTagPattern accepts Docker tag names used by release image publication.
var dockerTagPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)

type metadataConfig struct {
	// Framework stores framework-level release metadata.
	Framework metadataFrameworkConfig `yaml:"framework"`
}

type metadataFrameworkConfig struct {
	// Version stores the framework release version baseline.
	Version string `yaml:"version"`
}

// runReleaseTagCheck verifies that a release tag matches framework.version.
func runReleaseTagCheck(_ context.Context, a *app, input commandInput) error {
	tag := releaseTagFromInput(a.env, input)
	version, err := loadFrameworkVersion(a.root, input.Get("metadata"))
	if err != nil {
		return err
	}
	if input.HasBool("print_version") {
		if err = validateFrameworkReleaseVersion(version); err != nil {
			return err
		}
		fmt.Fprintln(a.stdout, version)
		return nil
	}
	if err = validateReleaseTagVersion(tag, version); err != nil {
		return err
	}
	fmt.Fprintf(a.stdout, "Release tag %s matches framework.version in metadata.yaml\n", tag)
	return nil
}

// releaseTagFromInput resolves the release tag from parameters or GitHub env.
func releaseTagFromInput(env []string, input commandInput) string {
	if tag := strings.TrimSpace(input.Get("tag")); tag != "" {
		return tag
	}
	return strings.TrimSpace(envValue(env, "GITHUB_REF_NAME"))
}

// loadFrameworkVersion reads framework.version from metadata.yaml.
func loadFrameworkVersion(root string, override string) (string, error) {
	metadataPath := strings.TrimSpace(override)
	if metadataPath == "" {
		metadataPath = filepath.Join("apps", "lina-core", "manifest", "config", "metadata.yaml")
	}
	if !filepath.IsAbs(metadataPath) {
		metadataPath = filepath.Join(root, metadataPath)
	}

	content, err := os.ReadFile(metadataPath)
	if err != nil {
		return "", fmt.Errorf("read metadata %s: %w", metadataPath, err)
	}
	var cfg metadataConfig
	if err = yaml.Unmarshal(content, &cfg); err != nil {
		return "", fmt.Errorf("parse metadata %s: %w", metadataPath, err)
	}
	version := strings.TrimSpace(cfg.Framework.Version)
	if version == "" {
		return "", fmt.Errorf("metadata framework.version is empty in %s", metadataPath)
	}
	return version, nil
}

// validateReleaseTagVersion checks equality and release tag compatibility.
func validateReleaseTagVersion(tag string, version string) error {
	tag = strings.TrimSpace(tag)
	version = strings.TrimSpace(version)
	if err := validateFrameworkReleaseVersion(version); err != nil {
		return err
	}
	if tag == "" {
		return fmt.Errorf("release tag is empty; pass tag=<version> or run from a tag workflow with GITHUB_REF_NAME")
	}
	if tag != version {
		return fmt.Errorf("release tag %q must equal metadata framework.version %q", tag, version)
	}
	return nil
}

// validateFrameworkReleaseVersion checks release metadata version syntax.
func validateFrameworkReleaseVersion(version string) error {
	version = strings.TrimSpace(version)
	if version == "" {
		return fmt.Errorf("metadata framework.version is empty")
	}
	if !releaseVersionPattern.MatchString(version) {
		return fmt.Errorf("metadata framework.version %q must match vMAJOR.MINOR.PATCH or vMAJOR.MINOR.PATCH-prerelease", version)
	}
	if !dockerTagPattern.MatchString(version) {
		return fmt.Errorf("metadata framework.version %q cannot be used as a Docker image tag", version)
	}
	return nil
}
