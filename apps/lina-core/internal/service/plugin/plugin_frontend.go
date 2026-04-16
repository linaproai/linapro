// This file exposes hosted frontend asset methods on the root plugin facade.

package plugin

import "context"

// PrewarmRuntimeFrontendBundles preloads frontend bundles for enabled dynamic plugins.
func (s *serviceImpl) PrewarmRuntimeFrontendBundles(ctx context.Context) error {
	return s.frontendSvc.PrewarmRuntimeFrontendBundles(ctx)
}

// ResolveRuntimeFrontendAsset resolves one frontend asset for a dynamic plugin.
func (s *serviceImpl) ResolveRuntimeFrontendAsset(
	ctx context.Context,
	pluginID string,
	version string,
	relativePath string,
) (*RuntimeFrontendAssetOutput, error) {
	return s.frontendSvc.ResolveRuntimeFrontendAsset(ctx, pluginID, version, relativePath)
}

// BuildRuntimeFrontendPublicBaseURL returns the public base URL for a plugin's hosted frontend assets.
func (s *serviceImpl) BuildRuntimeFrontendPublicBaseURL(pluginID string, version string) string {
	return s.frontendSvc.BuildRuntimeFrontendPublicBaseURL(pluginID, version)
}
