// This file exposes runtime and dynamic-route facade methods.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"

	"lina-core/internal/service/plugin/internal/openapi"
)

// StartRuntimeReconciler starts the background reconciler loop for dynamic plugins.
func (s *serviceImpl) StartRuntimeReconciler(ctx context.Context) {
	s.runtimeSvc.StartRuntimeReconciler(ctx)
}

// ReconcileRuntimePlugins runs one reconciliation pass for all dynamic plugins.
func (s *serviceImpl) ReconcileRuntimePlugins(ctx context.Context) error {
	return s.runtimeSvc.ReconcileRuntimePlugins(ctx)
}

// ListRuntimeStates returns public plugin runtime states for shell slot rendering.
func (s *serviceImpl) ListRuntimeStates(ctx context.Context) (*RuntimeStateListOutput, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	return s.runtimeSvc.ListRuntimeStates(ctx)
}

// PrewarmRuntimeFrontendBundles preloads frontend bundles for enabled dynamic plugins.
func (s *serviceImpl) PrewarmRuntimeFrontendBundles(ctx context.Context) error {
	readCtx, err := s.storeSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return err
	}
	return s.frontendSvc.PrewarmRuntimeFrontendBundles(readCtx)
}

// ResolveRuntimeFrontendAsset resolves one frontend asset for a dynamic plugin.
func (s *serviceImpl) ResolveRuntimeFrontendAsset(
	ctx context.Context,
	pluginID string,
	version string,
	relativePath string,
) (*RuntimeFrontendAssetOutput, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	if !s.integrationSvc.IsInstalledEnabledForTenant(ctx, pluginID) {
		return nil, errPluginPublicAssetNotFound(pluginID)
	}
	return s.frontendSvc.ResolveRuntimeFrontendAsset(ctx, pluginID, version, relativePath)
}

// BuildRuntimeFrontendPublicBaseURL returns the public base URL for a plugin's hosted frontend assets.
func (s *serviceImpl) BuildRuntimeFrontendPublicBaseURL(pluginID string, version string) string {
	return s.frontendSvc.BuildRuntimeFrontendPublicBaseURL(pluginID, version)
}

// errPluginPublicAssetNotFound returns a deliberately generic not-found error
// so disabled or tenant-unavailable plugins do not reveal public asset presence.
func errPluginPublicAssetNotFound(pluginID string) error {
	return gerror.Newf("plugin public asset is not available: %s", pluginID)
}

// ProjectDynamicRoutesToOpenAPI projects dynamic routes into the host OpenAPI paths.
func (s *serviceImpl) ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return err
	}
	return s.openapiSvc.ProjectDynamicRoutesToOpenAPI(ctx, paths)
}

// CurrentRevision returns the plugin-runtime cache revision used by derived
// read-model caches that are constructed below the root facade.
func (s *serviceImpl) CurrentRevision(ctx context.Context) (int64, error) {
	if s == nil || s.runtimeCacheRevisionCtrl == nil {
		return 0, nil
	}
	return s.runtimeCacheRevisionCtrl.CurrentRevision(ctx)
}

// BuildDynamicRoutePublicPath returns the host-visible public path for one
// dynamic plugin route contract.
func BuildDynamicRoutePublicPath(pluginID string, routePath string) string {
	return openapi.BuildRoutePublicPath(pluginID, routePath)
}

// UploadDynamicPackage validates and stores a runtime WASM package.
func (s *serviceImpl) UploadDynamicPackage(ctx context.Context, in *DynamicUploadInput) (*DynamicUploadOutput, error) {
	if err := s.ensurePlatformGovernance(ctx); err != nil {
		return nil, err
	}
	out, err := s.runtimeSvc.UploadDynamicPackage(ctx, in)
	if err != nil {
		return nil, err
	}
	if _, err = s.publishPluginChange(ctx, pluginChangePublishInput{
		pluginID:   out.Id,
		pluginType: out.Type,
		reason:     "dynamic_package_uploaded",
	}); err != nil {
		return nil, err
	}
	return out, nil
}

// PrepareDynamicRouteMiddleware prepares dynamic route state before the main handler.
func (s *serviceImpl) PrepareDynamicRouteMiddleware(r *ghttp.Request) {
	if r != nil {
		s.ensureRuntimeCacheFreshBestEffort(r.Context(), "prepare_dynamic_route")
	}
	s.runtimeSvc.PrepareDynamicRouteMiddleware(r)
}

// AuthenticateDynamicRouteMiddleware authenticates JWT tokens for dynamic routes.
func (s *serviceImpl) AuthenticateDynamicRouteMiddleware(r *ghttp.Request) {
	s.runtimeSvc.AuthenticateDynamicRouteMiddleware(r)
}

// RegisterDynamicRouteDispatcher binds the dynamic route catch-all handler to the group.
func (s *serviceImpl) RegisterDynamicRouteDispatcher(group *ghttp.RouterGroup) {
	s.runtimeSvc.RegisterDynamicRouteDispatcher(group)
}
