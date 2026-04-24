// This file provides semantic-version parsing and comparison helpers for
// framework-upgrade planning.

package frameworkupgrade

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// semanticVersionPattern validates simple semver values like v1.2.3 and optional prerelease suffixes.
var semanticVersionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?$`)

// semanticVersion stores comparable parts from one semver string.
type semanticVersion struct {
	Major      int    // Major stores the major version number.
	Minor      int    // Minor stores the minor version number.
	Patch      int    // Patch stores the patch version number.
	PreRelease string // PreRelease stores the optional prerelease suffix.
	Raw        string // Raw stores the original normalized version string.
}

// parseSemanticVersion parses one semver string into comparable numeric parts.
func parseSemanticVersion(value string) (*semanticVersion, error) {
	trimmed := strings.TrimSpace(value)
	match := semanticVersionPattern.FindStringSubmatch(trimmed)
	if len(match) < 4 {
		return nil, gerror.Newf("版本号必须使用 semver 语义化格式: %s", value)
	}

	major, err := strconv.Atoi(match[1])
	if err != nil {
		return nil, gerror.Wrapf(err, "解析 major 版本号失败: %s", value)
	}
	minor, err := strconv.Atoi(match[2])
	if err != nil {
		return nil, gerror.Wrapf(err, "解析 minor 版本号失败: %s", value)
	}
	patch, err := strconv.Atoi(match[3])
	if err != nil {
		return nil, gerror.Wrapf(err, "解析 patch 版本号失败: %s", value)
	}

	return &semanticVersion{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: strings.TrimSpace(match[4]),
		Raw:        trimmed,
	}, nil
}

// compareSemanticVersions compares two semver strings.
func compareSemanticVersions(left string, right string) (int, error) {
	leftVersion, err := parseSemanticVersion(left)
	if err != nil {
		return 0, err
	}
	rightVersion, err := parseSemanticVersion(right)
	if err != nil {
		return 0, err
	}
	return compareParsedSemanticVersions(leftVersion, rightVersion), nil
}

// compareParsedSemanticVersions compares two parsed semantic versions.
func compareParsedSemanticVersions(left *semanticVersion, right *semanticVersion) int {
	if left == nil && right == nil {
		return 0
	}
	if left == nil {
		return -1
	}
	if right == nil {
		return 1
	}
	if left.Major != right.Major {
		if left.Major > right.Major {
			return 1
		}
		return -1
	}
	if left.Minor != right.Minor {
		if left.Minor > right.Minor {
			return 1
		}
		return -1
	}
	if left.Patch != right.Patch {
		if left.Patch > right.Patch {
			return 1
		}
		return -1
	}
	if left.PreRelease == right.PreRelease {
		return 0
	}
	if left.PreRelease == "" {
		return 1
	}
	if right.PreRelease == "" {
		return -1
	}
	return strings.Compare(left.PreRelease, right.PreRelease)
}

// sortSemanticVersionsDesc sorts version strings from newest to oldest.
func sortSemanticVersionsDesc(values []string) ([]string, error) {
	versions := make([]*semanticVersion, 0, len(values))
	for _, item := range values {
		parsed, err := parseSemanticVersion(item)
		if err != nil {
			return nil, err
		}
		versions = append(versions, parsed)
	}
	sort.SliceStable(versions, func(i int, j int) bool {
		return compareParsedSemanticVersions(versions[i], versions[j]) > 0
	})

	items := make([]string, 0, len(versions))
	for _, item := range versions {
		items = append(items, item.Raw)
	}
	return items, nil
}
