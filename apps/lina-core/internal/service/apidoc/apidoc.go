// Package apidoc builds the host-managed OpenAPI document that powers the
// system API documentation page.
package apidoc

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"

	bizctxsvc "lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/pkg/pluginhost"
)

// ConfigProvider provides host OpenAPI metadata configuration.
type ConfigProvider interface {
	// GetOpenApi returns the current OpenAPI document metadata.
	GetOpenApi(ctx context.Context) *configsvc.OpenApiConfig
}

// PluginRouteProvider provides plugin route ownership and OpenAPI projection inputs.
type PluginRouteProvider interface {
	// ListSourceRouteBindings returns source-plugin route bindings captured during registration.
	ListSourceRouteBindings() []pluginhost.SourceRouteBinding
	// IsEnabled reports whether the given plugin is currently enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
	// ProjectDynamicRoutesToOpenAPI projects enabled dynamic-plugin routes into the OpenAPI paths.
	ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error
}

// Service defines the apidoc service contract.
type Service interface {
	// Build builds one host-managed OpenAPI document from the current route table
	// and current plugin enablement state.
	Build(ctx context.Context, server *ghttp.Server) (*goai.OpenApiV3, error)
	// ResolveRouteText resolves one route's localized module tag and operation
	// summary from the dedicated apidoc i18n catalog.
	ResolveRouteText(ctx context.Context, input RouteTextInput) RouteTextOutput
	// FindRouteTitleOperationKeys finds operation key bases whose localized
	// module tag contains the given keyword in the current request locale.
	FindRouteTitleOperationKeys(ctx context.Context, keyword string) []string
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	configSvc ConfigProvider
	bizCtxSvc bizctxsvc.Service
	i18nSvc   i18nsvc.Service
	pluginSvc PluginRouteProvider
}

// New creates and returns a new apidoc Service.
func New(configSvc ConfigProvider, pluginSvc PluginRouteProvider) Service {
	return &serviceImpl{
		configSvc: configSvc,
		bizCtxSvc: bizctxsvc.New(),
		i18nSvc:   i18nsvc.New(),
		pluginSvc: pluginSvc,
	}
}
