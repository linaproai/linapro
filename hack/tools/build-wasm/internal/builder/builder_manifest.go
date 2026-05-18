// This file loads the dynamic plugin manifest and validates manifest-level
// metadata shared by the standalone wasm builder flow.

package builder

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"lina-core/pkg/pluginbridge"
)

func validateRuntimeBuildManifest(manifest *pluginManifest, manifestPath string) error {
	if manifest == nil {
		return fmt.Errorf("dynamic plugin manifest cannot be nil")
	}
	if strings.TrimSpace(manifest.ID) == "" {
		return fmt.Errorf("dynamic plugin manifest missing id: %s", manifestPath)
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return fmt.Errorf("dynamic plugin manifest missing name: %s", manifestPath)
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return fmt.Errorf("dynamic plugin manifest missing version: %s", manifestPath)
	}
	manifest.Type = strings.ToLower(strings.TrimSpace(manifest.Type))
	if manifest.Type != pluginTypeDynamic {
		return fmt.Errorf("dynamic sample manifest type must be dynamic: %s", manifestPath)
	}
	manifest.ScopeNature = strings.ToLower(strings.TrimSpace(manifest.ScopeNature))
	if manifest.ScopeNature == "" {
		manifest.ScopeNature = pluginScopeNatureTenantAware
	}
	if manifest.ScopeNature != pluginScopeNaturePlatformOnly &&
		manifest.ScopeNature != pluginScopeNatureTenantAware {
		return fmt.Errorf("dynamic plugin scope_nature only supports platform_only/tenant_aware: %s", manifest.ScopeNature)
	}
	if manifest.SupportsMultiTenant == nil {
		return fmt.Errorf("dynamic plugin manifest missing supports_multi_tenant: %s", manifestPath)
	}
	if manifest.ScopeNature == pluginScopeNaturePlatformOnly && *manifest.SupportsMultiTenant {
		return fmt.Errorf("dynamic plugin supports_multi_tenant cannot be true when scope_nature is platform_only")
	}
	manifest.DefaultInstallMode = strings.ToLower(strings.TrimSpace(manifest.DefaultInstallMode))
	if manifest.ScopeNature == pluginScopeNaturePlatformOnly || !*manifest.SupportsMultiTenant {
		manifest.DefaultInstallMode = pluginInstallModeGlobal
	} else if manifest.DefaultInstallMode == "" {
		manifest.DefaultInstallMode = pluginInstallModeTenantScoped
	} else if manifest.DefaultInstallMode != pluginInstallModeGlobal &&
		manifest.DefaultInstallMode != pluginInstallModeTenantScoped {
		return fmt.Errorf("dynamic plugin default_install_mode only supports global/tenant_scoped: %s", manifest.DefaultInstallMode)
	}
	if !pluginManifestIDPattern.MatchString(manifest.ID) {
		return fmt.Errorf("dynamic plugin id must use kebab-case: %s", manifest.ID)
	}
	if err := validateSemanticVersion(manifest.Version); err != nil {
		return fmt.Errorf("dynamic plugin version is invalid: %w", err)
	}
	if err := validateDependencySpec(manifest.ID, manifest.Dependencies); err != nil {
		return fmt.Errorf("dynamic plugin dependencies invalid: %w", err)
	}
	manifest.Capabilities = pluginbridge.NormalizeCapabilities(manifest.Capabilities)
	if len(manifest.Capabilities) > 0 {
		return fmt.Errorf(
			"dynamic plugin manifest no longer supports top-level capabilities; please keep only hostServices declarations (found: %s)",
			strings.Join(manifest.Capabilities, ", "),
		)
	}
	if err := pluginbridge.ValidateHostServiceSpecs(manifest.HostServices); err != nil {
		return fmt.Errorf("dynamic plugin hostServices invalid: %w", err)
	}
	hostServices, err := pluginbridge.NormalizeHostServiceSpecs(manifest.HostServices)
	if err != nil {
		return fmt.Errorf("dynamic plugin hostServices normalization failed: %w", err)
	}
	manifest.HostServices = hostServices
	return nil
}

var buildDependencyVersionConstraintPattern = regexp.MustCompile(`^(>=|<=|>|<|=)?v?\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$`)

func validateDependencySpec(pluginID string, spec *dependencySpec) error {
	if spec == nil {
		return nil
	}
	if spec.Framework != nil {
		spec.Framework.Version = strings.TrimSpace(spec.Framework.Version)
		if spec.Framework.Version != "" {
			if err := validateSemanticVersionRange(spec.Framework.Version); err != nil {
				return fmt.Errorf("framework version is invalid: %w", err)
			}
		}
	}
	seen := make(map[string]struct{}, len(spec.Plugins))
	for index, dependency := range spec.Plugins {
		if dependency == nil {
			return fmt.Errorf("dependency %d cannot be nil", index+1)
		}
		dependency.ID = strings.TrimSpace(dependency.ID)
		dependency.Version = strings.TrimSpace(dependency.Version)
		dependency.Install = normalizeDependencyInstallModeForValidation(dependency.Install)
		if dependency.Required == nil {
			required := true
			dependency.Required = &required
		}
		if dependency.ID == "" {
			return fmt.Errorf("dependency %d is missing id", index+1)
		}
		if !pluginManifestIDPattern.MatchString(dependency.ID) {
			return fmt.Errorf("dependency id must use kebab-case: %s", dependency.ID)
		}
		if dependency.ID == pluginID {
			return fmt.Errorf("plugin cannot depend on itself: %s", pluginID)
		}
		if _, ok := seen[dependency.ID]; ok {
			return fmt.Errorf("duplicate dependency: %s", dependency.ID)
		}
		seen[dependency.ID] = struct{}{}
		if dependency.Version != "" {
			if err := validateSemanticVersionRange(dependency.Version); err != nil {
				return fmt.Errorf("dependency %s version is invalid: %w", dependency.ID, err)
			}
		}
		if dependency.Install == "" {
			return fmt.Errorf("dependency %s install only supports manual/auto", dependency.ID)
		}
	}
	return nil
}

func validateSemanticVersionRange(value string) error {
	tokens := strings.Fields(strings.TrimSpace(value))
	if len(tokens) == 0 {
		return fmt.Errorf("version range cannot be empty")
	}
	for _, token := range tokens {
		if !buildDependencyVersionConstraintPattern.MatchString(token) {
			return fmt.Errorf("version range token must use semver comparison format: %s", token)
		}
		if err := validateSemanticVersion(trimVersionConstraintOperator(token)); err != nil {
			return err
		}
	}
	return nil
}

func normalizeDependencyInstallModeForValidation(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	switch trimmed {
	case "", "manual":
		return "manual"
	case "auto":
		return "auto"
	default:
		return trimmed
	}
}

func trimVersionConstraintOperator(token string) string {
	token = strings.TrimSpace(token)
	for _, operator := range []string{">=", "<=", ">", "<", "="} {
		if strings.HasPrefix(token, operator) {
			return strings.TrimSpace(strings.TrimPrefix(token, operator))
		}
	}
	return token
}

func loadYAMLFile(filePath string, target interface{}) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if len(content) == 0 {
		return fmt.Errorf("yaml file is empty: %s", filePath)
	}
	if err = yaml.Unmarshal(content, target); err != nil {
		return fmt.Errorf("failed to parse yaml file %s: %w", filePath, err)
	}
	return nil
}

func validateSemanticVersion(value string) error {
	match := pluginManifestSemverPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) < 4 {
		return fmt.Errorf("version must use semver format: %s", value)
	}

	for _, raw := range match[1:4] {
		if _, err := strconv.Atoi(raw); err != nil {
			return err
		}
	}
	return nil
}
