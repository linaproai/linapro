// This file re-exports the install-mock-data context helpers from the catalog
// package so the plugin facade can attach and read the flag using local helper
// names while leaving the canonical implementation in one shared location.

package plugin

import (
	"context"

	"lina-core/internal/service/plugin/internal/catalog"
)

// withInstallMockData decorates ctx with the operator's install-mock-data
// opt-in flag. Backed by catalog.WithInstallMockData so source-plugin and
// dynamic-plugin install paths share the same key.
func withInstallMockData(ctx context.Context, enable bool) context.Context {
	return catalog.WithInstallMockData(ctx, enable)
}

// shouldInstallMockData reports whether the current request opted into mock
// data. Backed by catalog.ShouldInstallMockData.
func shouldInstallMockData(ctx context.Context) bool {
	return catalog.ShouldInstallMockData(ctx)
}
