// This file provides manifest-id and semantic-version validation helpers used
// while scanning plugin manifests.

package catalog

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Shared manifest validation regexes used during plugin catalog scanning.
var (
	// ManifestIDPattern is the allowed pattern for plugin IDs (kebab-case).
	ManifestIDPattern     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	manifestSemverPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?$`)
)

// semanticVersion provides lightweight semantic-version validation for plugin manifests.
type semanticVersion struct {
	Major int
	Minor int
	Patch int
}

// ValidateManifestSemanticVersion validates semantic version strings used by plugin.yaml.
func ValidateManifestSemanticVersion(value string) error {
	_, err := parseSemanticVersion(value)
	return err
}

// parseSemanticVersion parses a manifest semantic version into comparable parts.
func parseSemanticVersion(value string) (*semanticVersion, error) {
	match := manifestSemverPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) < 4 {
		return nil, gerror.Newf("version must use semver format: %s", value)
	}

	major, err := strconv.Atoi(match[1])
	if err != nil {
		return nil, err
	}
	minor, err := strconv.Atoi(match[2])
	if err != nil {
		return nil, err
	}
	patch, err := strconv.Atoi(match[3])
	if err != nil {
		return nil, err
	}

	return &semanticVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}
