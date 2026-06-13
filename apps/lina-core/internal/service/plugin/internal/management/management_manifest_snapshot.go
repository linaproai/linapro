// This file owns request-local manifest snapshots shared by list projection,
// dependency checks, and lifecycle install/status flows.

package management

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
)

// manifestSnapshotContextKey stores one request-local manifest discovery result.
type manifestSnapshotContextKey struct{}

// WithManifestSnapshot stores one already-scanned manifest list in context so
// dependency checks inside the same list build do not rescan source plugins and
// dynamic artifacts for every plugin row.
func WithManifestSnapshot(ctx context.Context, manifests []*catalog.Manifest) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if ManifestSnapshotFromContext(ctx) != nil {
		return ctx
	}
	return context.WithValue(ctx, manifestSnapshotContextKey{}, CloneManifestSlice(manifests))
}

// ManifestSnapshotFromContext returns the request-local manifest list, if set.
func ManifestSnapshotFromContext(ctx context.Context) []*catalog.Manifest {
	if ctx == nil {
		return nil
	}
	manifests, ok := ctx.Value(manifestSnapshotContextKey{}).([]*catalog.Manifest)
	if !ok || manifests == nil {
		return nil
	}
	return CloneManifestSlice(manifests)
}

// ManifestByIDFromContext returns a manifest from the request-local discovery
// snapshot without triggering another scan.
func ManifestByIDFromContext(ctx context.Context, pluginID string) *catalog.Manifest {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil
	}
	for _, manifest := range ManifestSnapshotFromContext(ctx) {
		if manifest != nil && strings.TrimSpace(manifest.ID) == normalizedPluginID {
			return manifest
		}
	}
	return nil
}

// CloneManifestSlice copies the manifest slice header so callers cannot mutate
// the request-local list ordering.
func CloneManifestSlice(in []*catalog.Manifest) []*catalog.Manifest {
	if in == nil {
		return nil
	}
	out := make([]*catalog.Manifest, len(in))
	copy(out, in)
	return out
}
