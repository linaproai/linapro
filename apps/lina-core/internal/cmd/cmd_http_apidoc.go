// This file binds the host-managed OpenAPI document endpoint.

package cmd

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"

	"lina-core/internal/model"
	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/bizctx"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/pkg/logger"
)

// bindHostedOpenAPIDocs disables the GoFrame built-in OpenAPI and Swagger
// endpoints, then binds the host-managed OpenAPI JSON handler at the configured path.
func bindHostedOpenAPIDocs(
	_ context.Context,
	server *ghttp.Server,
	apiDocSvc apidoc.Service,
	apiDocPath string,
	apiDocI18nSvc i18nsvc.Service,
	bizCtxSvc bizctx.Service,
) {
	if server == nil {
		return
	}

	server.SetOpenApiPath("")
	server.SetSwaggerPath("")

	apiDocPath = strings.TrimSpace(apiDocPath)
	if apiDocPath == "" || apiDocSvc == nil {
		return
	}

	server.BindHandler(apiDocPath, func(r *ghttp.Request) {
		bizCtxSvc.Init(r, &model.Context{})
		locale := apiDocI18nSvc.ResolveRequestLocale(r)
		r.SetCtx(gi18n.WithLanguage(r.Context(), locale))
		bizCtxSvc.SetLocale(r.Context(), locale)
		r.Response.Header().Set("Content-Language", locale)

		document, err := apiDocSvc.Build(r.Context(), server)
		if err != nil {
			logger.Warningf(r.Context(), "build hosted OpenAPI document failed: %v", err)
			r.Response.WriteStatus(http.StatusInternalServerError)
			r.Response.Write("build hosted OpenAPI document failed")
			r.ExitAll()
			return
		}
		applyOpenAPIRequestOrigin(r, document)
		r.Response.WriteJson(document)
		r.ExitAll()
	})
}

// applyOpenAPIRequestOrigin updates OpenAPI servers so documentation clients
// use the effective backend origin that served the document.
func applyOpenAPIRequestOrigin(r *ghttp.Request, document *goai.OpenApiV3) {
	if r == nil || document == nil {
		return
	}
	origin := buildOpenAPIRequestOrigin(r)
	if origin == "" {
		return
	}

	description := ""
	if document.Servers != nil && len(*document.Servers) > 0 {
		description = (*document.Servers)[0].Description
	}
	document.Servers = &goai.Servers{
		{
			URL:         origin,
			Description: description,
		},
	}
}

// buildOpenAPIRequestOrigin returns the scheme and host portion of the current
// OpenAPI document request while preserving the request port.
func buildOpenAPIRequestOrigin(r *ghttp.Request) string {
	if r == nil || r.Request == nil {
		return ""
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return ""
	}
	scheme := strings.TrimSpace(r.GetSchema())
	if scheme == "" {
		scheme = "http"
	}
	return scheme + "://" + host
}
