// This file validates dependency declarations required by owner-aware dynamic
// host services. It keeps cross-plugin dependency policy at the manifest
// boundary where both hostServices and dependencies.plugins are visible.

package catalog

import (
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// ValidateOwnerHostServiceDependencies ensures every owner-aware host service
// has a matching hard plugin dependency with a version range.
func ValidateOwnerHostServiceDependencies(manifest *Manifest) error {
	if manifest == nil || len(manifest.HostServices) == 0 {
		return nil
	}
	dependencies := pluginDependencyVersionMap(manifest.Dependencies)
	ownerSpecs := ownerHostServiceSpecs(manifest.HostServices)
	for _, spec := range ownerSpecs {
		owner := strings.TrimSpace(spec.Owner)
		if owner == "" {
			continue
		}
		version, ok := dependencies[owner]
		if !ok {
			return gerror.Newf(
				"plugin %s owner host service %s requires dependencies.plugins entry for owner plugin %s",
				manifest.ID,
				protocol.HostServiceSpecLabel(spec),
				owner,
			)
		}
		if strings.TrimSpace(version) == "" {
			return gerror.Newf(
				"plugin %s owner host service %s requires dependencies.plugins version range for owner plugin %s",
				manifest.ID,
				protocol.HostServiceSpecLabel(spec),
				owner,
			)
		}
	}
	return nil
}

func normalizeManifestHostServices(manifest *Manifest) error {
	if manifest == nil {
		return nil
	}
	hostServices, err := protocol.NormalizeHostServiceSpecsForPlugin(manifest.ID, manifest.HostServices)
	if err != nil {
		return err
	}
	manifest.HostServices = hostServices
	manifest.HostCapabilities = protocol.CapabilityMapFromHostServices(hostServices)
	return nil
}

func pluginDependencyVersionMap(spec *plugintypes.DependencySpec) map[string]string {
	result := make(map[string]string)
	if spec == nil {
		return result
	}
	for _, dependency := range spec.Plugins {
		if dependency == nil {
			continue
		}
		pluginID := strings.TrimSpace(dependency.ID)
		if pluginID == "" {
			continue
		}
		result[pluginID] = strings.TrimSpace(dependency.Version)
	}
	return result
}

func ownerHostServiceSpecs(specs []*protocol.HostServiceSpec) []*protocol.HostServiceSpec {
	result := make([]*protocol.HostServiceSpec, 0, len(specs))
	for _, spec := range specs {
		if spec == nil || strings.TrimSpace(spec.Owner) == "" {
			continue
		}
		result = append(result, spec)
	}
	sort.Slice(result, func(i int, j int) bool {
		return protocol.HostServiceSpecLabel(result[i]) < protocol.HostServiceSpecLabel(result[j])
	})
	return result
}
