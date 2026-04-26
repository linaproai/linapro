// This file implements host-managed OpenAPI document construction from the
// current host route table and plugin route projections.

package apidoc

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"

	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

// Build builds one host-managed OpenAPI document from the current route table
// and current plugin enablement state.
func (s *serviceImpl) Build(ctx context.Context, server *ghttp.Server) (*goai.OpenApiV3, error) {
	if server == nil {
		return nil, gerror.New("apidoc: host server is nil")
	}

	document := s.newDocument(ctx)
	sourceRouteBindings := s.listSourceRouteBindings()
	sourceRouteKeySet := buildSourceRouteKeySet(sourceRouteBindings)

	if err := s.addHostStaticRoutes(document, server, sourceRouteKeySet); err != nil {
		return nil, err
	}
	s.addEnabledSourceRoutes(ctx, document, sourceRouteBindings)
	if s.pluginSvc != nil {
		if err := s.pluginSvc.ProjectDynamicRoutesToOpenAPI(ctx, document.Paths); err != nil {
			return nil, err
		}
	}
	s.localizeDocument(ctx, document)
	return document, nil
}

// newDocument creates the baseline host-managed OpenAPI document and applies
// configured document metadata and shared security defaults.
func (s *serviceImpl) newDocument(ctx context.Context) *goai.OpenApiV3 {
	document := goai.New()
	if document.Paths == nil {
		document.Paths = goai.Paths{}
	}

	document.Config.CommonResponse = ghttp.DefaultHandlerResponse{}
	document.Config.CommonResponseDataField = "Data"
	document.Components.SecuritySchemes = goai.SecuritySchemes{
		"BearerAuth": goai.SecuritySchemeRef{
			Value: &goai.SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
				Description:  "JWT Bearer Token Authentication",
				In:           "header",
				Name:         "Authorization",
			},
		},
	}
	document.Security = &goai.SecurityRequirements{
		{"BearerAuth": {}},
	}

	if s == nil || s.configSvc == nil {
		return document
	}
	oaiCfg := s.configSvc.GetOpenApi(ctx)
	if oaiCfg == nil {
		return document
	}
	document.Info.Title = oaiCfg.Title
	document.Info.Description = oaiCfg.Description
	document.Info.Version = oaiCfg.Version
	if strings.TrimSpace(oaiCfg.ServerUrl) != "" {
		document.Servers = &goai.Servers{
			{
				URL:         oaiCfg.ServerUrl,
				Description: oaiCfg.ServerDescription,
			},
		}
	}
	return document
}

// addHostStaticRoutes projects host-owned strict routes that are not shadowed
// by source-plugin bindings into the output OpenAPI document.
func (s *serviceImpl) addHostStaticRoutes(
	document *goai.OpenApiV3,
	server *ghttp.Server,
	sourceRouteKeySet map[string]struct{},
) error {
	if document == nil || server == nil {
		return nil
	}
	for _, route := range server.GetRoutes() {
		if !shouldIncludeHostStaticRoute(route, sourceRouteKeySet) {
			continue
		}
		if err := addHandlerRouteToOpenAPI(
			document, route.Route, route.Method, route.Handler.Info.Value.Interface(),
		); err != nil {
			return err
		}
	}
	return nil
}

// addEnabledSourceRoutes projects documentable source-plugin routes for the
// plugins that are currently enabled.
func (s *serviceImpl) addEnabledSourceRoutes(
	ctx context.Context,
	document *goai.OpenApiV3,
	bindings []pluginhost.SourceRouteBinding,
) {
	if document == nil || len(bindings) == 0 {
		return
	}

	projectedRouteSet := make(map[string]struct{}, len(bindings))
	for _, binding := range bindings {
		if !binding.Documentable {
			continue
		}
		if s.pluginSvc != nil && !s.pluginSvc.IsEnabled(ctx, binding.PluginID) {
			continue
		}

		key := binding.Key()
		if _, ok := projectedRouteSet[key]; ok {
			continue
		}
		if err := addHandlerRouteToOpenAPI(document, binding.Path, binding.Method, binding.Handler); err != nil {
			logger.Warningf(
				ctx,
				"project source plugin route to OpenAPI failed plugin=%s method=%s path=%s err=%v",
				binding.PluginID,
				binding.Method,
				binding.Path,
				err,
			)
			continue
		}
		projectedRouteSet[key] = struct{}{}
	}
}

// listSourceRouteBindings reads the current source-plugin route binding snapshot
// from the plugin service when available.
func (s *serviceImpl) listSourceRouteBindings() []pluginhost.SourceRouteBinding {
	if s == nil || s.pluginSvc == nil {
		return nil
	}
	return s.pluginSvc.ListSourceRouteBindings()
}

// shouldIncludeHostStaticRoute reports whether the host route should stay in
// the document after removing plugin-owned strict-route duplicates.
func shouldIncludeHostStaticRoute(route ghttp.RouterItem, sourceRouteKeySet map[string]struct{}) bool {
	if route.Handler == nil || !route.Handler.Info.IsStrictRoute {
		return false
	}
	if _, ok := sourceRouteKeySet[buildRouteKey(route.Method, route.Route)]; ok {
		return false
	}
	return true
}

// addHandlerRouteToOpenAPI expands the handler's method set and registers it
// into the target OpenAPI document.
func addHandlerRouteToOpenAPI(
	document *goai.OpenApiV3,
	path string,
	method string,
	handler interface{},
) error {
	if document == nil {
		return nil
	}
	methods := expandOpenAPIMethods(method)
	for _, item := range methods {
		if err := document.Add(goai.AddInput{
			Path:   path,
			Method: item,
			Object: handler,
		}); err != nil {
			return err
		}
	}
	return nil
}

// buildSourceRouteKeySet builds one lookup set for source-plugin route keys.
func buildSourceRouteKeySet(bindings []pluginhost.SourceRouteBinding) map[string]struct{} {
	items := make(map[string]struct{}, len(bindings))
	for _, binding := range bindings {
		items[binding.Key()] = struct{}{}
	}
	return items
}

// buildRouteKey combines one method and path into the normalized route key used
// by host and plugin route de-duplication.
func buildRouteKey(method string, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + normalizeOpenAPIPath(path)
}

// normalizeOpenAPIPath canonicalizes an OpenAPI path for stable key comparison.
func normalizeOpenAPIPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || trimmed == "/" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/")
}

// expandOpenAPIMethods expands ALL or empty methods to the full supported HTTP
// method set used by GoFrame OpenAPI generation.
func expandOpenAPIMethods(method string) []string {
	normalized := strings.ToUpper(strings.TrimSpace(method))
	if normalized == "" || normalized == "ALL" {
		methods := ghttp.SupportedMethods()
		items := make([]string, 0, len(methods))
		for _, item := range methods {
			items = append(items, strings.ToUpper(strings.TrimSpace(item)))
		}
		return items
	}
	return []string{normalized}
}
