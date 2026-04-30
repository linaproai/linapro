// This file binds the host-managed OpenAPI document endpoint.

package cmd

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/net/ghttp"

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

	apiDocBizCtxSvc := bizctx.New()
	apiDocI18nSvc := i18nsvc.New()
	server.BindHandler(apiDocPath, func(r *ghttp.Request) {
		apiDocBizCtxSvc.Init(r, &model.Context{})
		locale := apiDocI18nSvc.ResolveRequestLocale(r)
		r.SetCtx(gi18n.WithLanguage(r.Context(), locale))
		apiDocBizCtxSvc.SetLocale(r.Context(), locale)
		r.Response.Header().Set("Content-Language", locale)

		document, err := apiDocSvc.Build(r.Context(), server)
		if err != nil {
			logger.Warningf(r.Context(), "build hosted OpenAPI document failed: %v", err)
			r.Response.WriteStatus(http.StatusInternalServerError)
			r.Response.Write("build hosted OpenAPI document failed")
			r.ExitAll()
			return
		}
		r.Response.WriteJson(document)
		r.ExitAll()
	})
}
