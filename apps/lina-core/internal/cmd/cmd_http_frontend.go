// This file serves embedded frontend assets and dynamic plugin frontend assets.

package cmd

import (
	"context"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/packed"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/logger"
)

// bindFrontendAssetRoutes registers the final catch-all static route for host
// frontend assets and SPA fallback after all API and plugin routes are bound.
func bindFrontendAssetRoutes(
	ctx context.Context,
	server *ghttp.Server,
	pluginSvc pluginsvc.Service,
) error {
	subFS, err := fs.Sub(packed.Files, "public")
	if err != nil {
		logger.Panicf(ctx, "load embedded frontend assets failed: %v", err)
		return err
	}
	fileServer := http.FileServer(http.FS(subFS))
	server.BindHandler("/*", newFrontendAssetHandler(subFS, fileServer, pluginSvc))
	return nil
}

// newFrontendAssetHandler creates the SPA/static-file handler used as the last
// route in the server so API and plugin routes get first chance to match.
func newFrontendAssetHandler(
	subFS fs.FS,
	fileServer http.Handler,
	pluginSvc pluginsvc.Service,
) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if serveRuntimePluginAsset(r, pluginSvc, path) {
			return
		}
		if serveEmbeddedFrontendAsset(r, subFS, fileServer, path) {
			return
		}
		serveSPAFallback(r, fileServer)
	}
}

// serveRuntimePluginAsset serves versioned dynamic plugin frontend assets when
// the request path belongs to the public plugin-asset namespace.
func serveRuntimePluginAsset(
	r *ghttp.Request,
	pluginSvc pluginsvc.Service,
	path string,
) bool {
	// Runtime plugin assets must be checked before the host falls back to the
	// embedded frontend bundle. They share the same public static entrypoint,
	// but plugin assets are governed by plugin ID, version, and enabled state.
	// If the host served the generic SPA assets first, a valid plugin asset URL
	// could be swallowed by the host fallback and bypass the runtime-specific
	// access rules that ResolveRuntimeFrontendAsset enforces.
	pluginID, version, assetPath, ok := parsePluginAssetRequestPath(path)
	if !ok {
		return false
	}
	out, resolveErr := pluginSvc.ResolveRuntimeFrontendAsset(
		r.Context(),
		pluginID,
		version,
		assetPath,
	)
	if resolveErr != nil {
		r.Response.WriteStatus(http.StatusNotFound)
		r.ExitAll()
		return true
	}
	r.Response.Header().Set("Content-Type", out.ContentType)
	r.Response.Write(out.Content)
	r.ExitAll()
	return true
}

// serveEmbeddedFrontendAsset serves one concrete embedded frontend file when
// it exists and lets callers fall through to the SPA fallback otherwise.
func serveEmbeddedFrontendAsset(
	r *ghttp.Request,
	subFS fs.FS,
	fileServer http.Handler,
	path string,
) bool {
	f, err := subFS.Open(path)
	if err != nil {
		return false
	}
	if closeErr := f.Close(); closeErr != nil {
		logger.Warningf(
			r.Context(),
			"close embedded frontend asset failed path=%s err=%v",
			path,
			closeErr,
		)
	}
	fileServer.ServeHTTP(r.Response.RawWriter(), r.Request)
	r.ExitAll()
	return true
}

// serveSPAFallback serves index.html for unmatched frontend routes so browser
// refreshes on client-side routes are handled by the Vue application.
func serveSPAFallback(r *ghttp.Request, fileServer http.Handler) {
	r.Request.URL.Path = "/index.html"
	fileServer.ServeHTTP(r.Response.RawWriter(), r.Request)
	r.ExitAll()
}

// parsePluginAssetRequestPath splits one public `/plugin-assets/...` request
// path into plugin identity, version, and relative asset path parts.
func parsePluginAssetRequestPath(path string) (
	pluginID string,
	version string,
	assetPath string,
	ok bool,
) {
	normalizedPath := strings.Trim(strings.TrimSpace(path), "/")
	if normalizedPath == "" {
		return "", "", "", false
	}

	pathParts := strings.Split(normalizedPath, "/")
	if len(pathParts) < 3 || pathParts[0] != "plugin-assets" {
		return "", "", "", false
	}
	if strings.TrimSpace(pathParts[1]) == "" || strings.TrimSpace(pathParts[2]) == "" {
		return "", "", "", false
	}

	pluginID = pathParts[1]
	version = pathParts[2]
	if len(pathParts) == 3 {
		return pluginID, version, "", true
	}
	return pluginID, version, strings.Join(pathParts[3:], "/"), true
}
