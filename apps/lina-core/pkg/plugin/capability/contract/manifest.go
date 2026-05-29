// This file defines the source-plugin visible manifest resource contract.

package contract

import "context"

// ManifestServiceFactory creates plugin-scoped manifest resource service views.
type ManifestServiceFactory interface {
	// ForPlugin returns a manifest resource service scoped to pluginID. Blank
	// plugin IDs return a service that rejects reads.
	ForPlugin(pluginID string) ManifestService
	// WithArtifactResources returns a new factory view that can use release-bound
	// artifact resources for pluginID. Paths are relative to manifest/.
	WithArtifactResources(pluginID string, resources map[string][]byte) ManifestServiceFactory
}

// ManifestService defines read-only access to one plugin's manifest resources.
type ManifestService interface {
	// Get returns one raw resource under the current plugin manifest
	// root. Paths are slash-separated and relative to manifest/.
	Get(ctx context.Context, path string) ([]byte, error)
	// Exists reports whether one allowed manifest resource exists under the
	// current plugin manifest root.
	Exists(ctx context.Context, path string) (bool, error)
	// Scan unmarshals the selected YAML resource, or the nested key inside it,
	// into target. Missing resources leave target unchanged.
	Scan(ctx context.Context, path string, key string, target any) error
}
