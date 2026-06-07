// This file projects validated dynamic artifact manifest resources into the
// per-execution views consumed by plugin config and manifest host services.

package runtime

import (
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	capabilityconfig "lina-core/pkg/plugin/capability/plugincap"
)

// buildArtifactDefaultConfig returns the active-release default config content
// from manifest/config/config.yaml. The template config.example.yaml is never
// exposed as runtime defaults.
func buildArtifactDefaultConfig(manifest *catalog.Manifest) []byte {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return nil
	}
	for _, resource := range manifest.RuntimeArtifact.ManifestResources {
		if resource == nil {
			continue
		}
		if strings.TrimSpace(resource.Path) == "manifest/config/"+capabilityconfig.RuntimeConfigFileName {
			return append([]byte(nil), resource.Content...)
		}
	}
	return nil
}

// buildArtifactManifestResources returns raw manifest resources keyed relative
// to manifest/. Dedicated config, SQL, and i18n pipelines decide how those
// resources take effect; this view only exposes their original bytes.
func buildArtifactManifestResources(manifest *catalog.Manifest) map[string][]byte {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return nil
	}
	resources := make(map[string][]byte)
	for _, resource := range manifest.RuntimeArtifact.ManifestResources {
		if resource == nil {
			continue
		}
		relativePath := strings.TrimPrefix(strings.TrimSpace(resource.Path), "manifest/")
		if relativePath == "" || relativePath == resource.Path {
			continue
		}
		resources[relativePath] = append([]byte(nil), resource.Content...)
	}
	if len(resources) == 0 {
		return nil
	}
	return resources
}
