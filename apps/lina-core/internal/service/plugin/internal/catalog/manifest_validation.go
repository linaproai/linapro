// This file provides repository-root, manifest-id, and semantic-version
// validation helpers used while scanning plugin manifests.

package catalog

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
)

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

// FindRepoRoot walks upward from the provided directory to locate the repository root.
func FindRepoRoot(startDir string) (string, error) {
	current, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	current = filepath.Clean(current)
	for depth := 0; depth < 8; depth++ {
		if gfile.Exists(filepath.Join(current, "apps", "lina-core", "go.mod")) &&
			gfile.Exists(filepath.Join(current, "apps", "lina-vben", "package.json")) {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", gerror.Newf("未找到仓库根目录，起始路径: %s", startDir)
}

// ValidateManifestSemanticVersion validates semantic version strings used by plugin.yaml.
func ValidateManifestSemanticVersion(value string) error {
	_, err := parseSemanticVersion(value)
	return err
}

func parseSemanticVersion(value string) (*semanticVersion, error) {
	match := manifestSemverPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) < 4 {
		return nil, gerror.Newf("版本号必须使用 semver 语义化格式: %s", value)
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
