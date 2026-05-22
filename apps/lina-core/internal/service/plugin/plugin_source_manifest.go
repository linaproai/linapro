// This file exposes source-plugin manifest discovery for governance callers
// that need the unified catalog parser without constructing the full plugin
// service graph or importing plugin/internal packages directly.

package plugin

import "lina-core/internal/service/plugin/internal/catalog"

// SourceManifest aliases the framework plugin manifest model discovered from
// registered source plugins.
type SourceManifest = catalog.Manifest

// ScanRegisteredSourceManifests returns registered source-plugin manifests
// without synchronizing registry, release, menu, permission, or cache state.
func ScanRegisteredSourceManifests() ([]*SourceManifest, error) {
	return catalog.New(nil).ScanEmbeddedSourceManifests()
}
