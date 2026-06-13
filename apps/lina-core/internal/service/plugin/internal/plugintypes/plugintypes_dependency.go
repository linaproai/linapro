// This file validates and normalizes plugin dependency declarations used by
// source and dynamic plugin manifests.

package plugintypes

import (
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

var dependencyVersionConstraintPattern = regexp.MustCompile(`^(>=|<=|>|<|=)?v?\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$`)

// DependencySpec defines the dependency constraints declared by one plugin.
type DependencySpec struct {
	// Framework declares LinaPro framework compatibility constraints.
	Framework *FrameworkDependencySpec `json:"framework,omitempty" yaml:"framework,omitempty"`
	// Plugins lists other plugins this plugin depends on.
	Plugins []*PluginDependencySpec `json:"plugins,omitempty" yaml:"plugins,omitempty"`
}

// FrameworkDependencySpec declares the compatible LinaPro framework version range.
type FrameworkDependencySpec struct {
	// Version is the semantic version range required by this plugin.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

// PluginDependencySpec declares one plugin-to-plugin dependency.
type PluginDependencySpec struct {
	// ID is the depended-on plugin ID.
	ID string `json:"id" yaml:"id"`
	// Version is the semantic version range required from the dependency.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

// NormalizeDependencySpec applies manifest defaults and trims dependency
// values in-place. Missing dependencies remain nil.
func NormalizeDependencySpec(spec *DependencySpec) {
	if spec == nil {
		return
	}
	if spec.Framework != nil {
		spec.Framework.Version = strings.TrimSpace(spec.Framework.Version)
		if spec.Framework.Version == "" {
			spec.Framework = nil
		}
	}

	plugins := make([]*PluginDependencySpec, 0, len(spec.Plugins))
	for _, dependency := range spec.Plugins {
		if dependency == nil {
			plugins = append(plugins, dependency)
			continue
		}
		dependency.ID = strings.TrimSpace(dependency.ID)
		dependency.Version = strings.TrimSpace(dependency.Version)
		plugins = append(plugins, dependency)
	}
	spec.Plugins = plugins
}

// ValidateDependencySpec validates one plugin dependency declaration and
// normalizes defaults in-place.
func ValidateDependencySpec(pluginID string, spec *DependencySpec) error {
	if spec == nil {
		return nil
	}
	NormalizeDependencySpec(spec)
	if spec.Framework != nil {
		if err := ValidateSemanticVersionRange(spec.Framework.Version); err != nil {
			return gerror.Wrapf(err, "plugin %s framework dependency version is invalid", pluginID)
		}
	}

	seen := make(map[string]struct{}, len(spec.Plugins))
	for index, dependency := range spec.Plugins {
		if dependency == nil {
			return gerror.Newf("plugin %s dependency %d cannot be nil", pluginID, index+1)
		}
		if dependency.ID == "" {
			return gerror.Newf("plugin %s dependency %d is missing id", pluginID, index+1)
		}
		if err := ValidatePluginID(dependency.ID); err != nil {
			return gerror.Wrapf(err, "plugin %s dependency id is invalid", pluginID)
		}
		if dependency.ID == pluginID {
			return gerror.Newf("plugin %s cannot depend on itself", pluginID)
		}
		if _, ok := seen[dependency.ID]; ok {
			return gerror.Newf("plugin %s declares duplicate dependency: %s", pluginID, dependency.ID)
		}
		seen[dependency.ID] = struct{}{}
		if strings.TrimSpace(dependency.Version) != "" {
			if err := ValidateSemanticVersionRange(dependency.Version); err != nil {
				return gerror.Wrapf(err, "plugin %s dependency %s version is invalid", pluginID, dependency.ID)
			}
		}
	}
	return nil
}

// MatchesSemanticVersionRange reports whether version satisfies a
// whitespace-separated semantic-version constraint expression.
func MatchesSemanticVersionRange(version string, value string) (bool, error) {
	if err := ValidateManifestSemanticVersion(version); err != nil {
		return false, err
	}
	if err := ValidateSemanticVersionRange(value); err != nil {
		return false, err
	}
	for _, token := range strings.Fields(strings.TrimSpace(value)) {
		operator, constraintVersion := splitVersionConstraint(token)
		compare, err := CompareSemanticVersions(version, constraintVersion)
		if err != nil {
			return false, err
		}
		if !semanticVersionConstraintMatches(compare, operator) {
			return false, nil
		}
	}
	return true, nil
}

// ValidateSemanticVersionRange validates a whitespace-separated semantic
// version constraint expression such as ">=0.6.0 <0.7.0".
func ValidateSemanticVersionRange(value string) error {
	tokens := strings.Fields(strings.TrimSpace(value))
	if len(tokens) == 0 {
		return gerror.New("version range cannot be empty")
	}
	for _, token := range tokens {
		if !dependencyVersionConstraintPattern.MatchString(token) {
			return gerror.Newf("version range token must use semver comparison format: %s", token)
		}
		version := trimVersionConstraintOperator(token)
		if err := ValidateManifestSemanticVersion(version); err != nil {
			return err
		}
	}
	return nil
}

// CloneDependencySpec deep-copies a dependency declaration so release snapshots
// and runtime projections do not alias mutable manifest state.
func CloneDependencySpec(spec *DependencySpec) *DependencySpec {
	if spec == nil {
		return nil
	}
	clone := &DependencySpec{}
	if spec.Framework != nil {
		clone.Framework = &FrameworkDependencySpec{Version: strings.TrimSpace(spec.Framework.Version)}
	}
	if len(spec.Plugins) > 0 {
		clone.Plugins = make([]*PluginDependencySpec, 0, len(spec.Plugins))
		for _, dependency := range spec.Plugins {
			if dependency == nil {
				clone.Plugins = append(clone.Plugins, nil)
				continue
			}
			clone.Plugins = append(clone.Plugins, &PluginDependencySpec{
				ID:      strings.TrimSpace(dependency.ID),
				Version: strings.TrimSpace(dependency.Version),
			})
		}
	}
	return clone
}

// splitVersionConstraint separates one range token into operator and version.
func splitVersionConstraint(token string) (string, string) {
	token = strings.TrimSpace(token)
	for _, operator := range []string{">=", "<=", ">", "<", "="} {
		if strings.HasPrefix(token, operator) {
			return operator, strings.TrimSpace(strings.TrimPrefix(token, operator))
		}
	}
	return "=", token
}

// semanticVersionConstraintMatches applies one comparison result to a
// normalized semantic-version range operator.
func semanticVersionConstraintMatches(compare int, operator string) bool {
	switch operator {
	case ">":
		return compare > 0
	case ">=":
		return compare >= 0
	case "<":
		return compare < 0
	case "<=":
		return compare <= 0
	default:
		return compare == 0
	}
}

// trimVersionConstraintOperator removes a leading comparison operator from one
// version constraint token and returns the raw semver value.
func trimVersionConstraintOperator(token string) string {
	token = strings.TrimSpace(token)
	for _, operator := range []string{">=", "<=", ">", "<", "="} {
		if strings.HasPrefix(token, operator) {
			return strings.TrimSpace(strings.TrimPrefix(token, operator))
		}
	}
	return token
}
