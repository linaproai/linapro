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

// sourceLifecycleInstallOptionsContextKey isolates source-plugin lifecycle
// metadata attached by the root install facade.
type sourceLifecycleInstallOptionsContextKey struct{}

// sourceLifecycleInstallOptions carries install metadata that is safe to expose
// to source-plugin lifecycle callbacks.
type sourceLifecycleInstallOptions struct {
	startupAutoEnable bool
}

// withSourceLifecycleInstallOptions decorates ctx with source lifecycle
// metadata for the current install request.
func withSourceLifecycleInstallOptions(ctx context.Context, options InstallOptions) context.Context {
	return context.WithValue(ctx, sourceLifecycleInstallOptionsContextKey{}, sourceLifecycleInstallOptions{
		startupAutoEnable: options.startupAutoEnable,
	})
}

// sourceLifecycleStartupAutoEnable reports whether the current source-plugin
// install was initiated by startup plugin.autoEnable.
func sourceLifecycleStartupAutoEnable(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	options, ok := ctx.Value(sourceLifecycleInstallOptionsContextKey{}).(sourceLifecycleInstallOptions)
	return ok && options.startupAutoEnable
}
