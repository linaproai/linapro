// This file serves embedded frontend assets and dynamic plugin frontend assets.

package httpstartup

import (
	"context"
	"io/fs"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/packed"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
)

const frontendDevServerURLEnv = "LINAPRO_FRONTEND_DEV_SERVER_URL"

// bindFrontendAssetRoutes registers the final frontend catch-all route after
// API and plugin routes are bound. The handler only serves `/x-assets` and the
// configured workspace base path; other unmatched paths return 404.
func bindFrontendAssetRoutes(
	ctx context.Context,
	server *ghttp.Server,
	pluginSvc pluginsvc.Service,
	workspaceBasePath string,
) error {
	subFS, err := fs.Sub(packed.Files, "public")
	if err != nil {
		logger.Errorf(ctx, "load embedded frontend assets failed: %v", err)
		return err
	}
	return bindFrontendAssetRoutesWithFS(server, pluginSvc, workspaceBasePath, subFS)
}

// bindFrontendAssetRoutesWithFS registers frontend routes against a caller
// supplied filesystem so tests can avoid depending on generated build assets.
func bindFrontendAssetRoutesWithFS(
	server *ghttp.Server,
	pluginSvc pluginsvc.Service,
	workspaceBasePath string,
	frontendFS fs.FS,
) error {
	devProxy, err := newFrontendDevServerProxy()
	if err != nil {
		return err
	}
	normalizedWorkspaceBasePath := normalizeWorkspaceRequestBasePath(workspaceBasePath)
	assetHandler := newFrontendAssetHandler(frontendFS, pluginSvc, devProxy, normalizedWorkspaceBasePath)
	hostedAssetURLPrefix := strings.TrimRight(pluginhost.HostedAssetURLPrefix, "/")
	server.BindHandler(hostedAssetURLPrefix, assetHandler)
	server.BindHandler(hostedAssetURLPrefix+"/*any", assetHandler)
	if normalizedWorkspaceBasePath == "/" {
		server.BindHandler("/", assetHandler)
	} else {
		server.BindHandler(normalizedWorkspaceBasePath, assetHandler)
		server.BindHandler(normalizedWorkspaceBasePath+"/*any", assetHandler)
	}
	server.BindHandler("/{entry}", assetHandler)
	server.BindHandler("/{entry}/*any", assetHandler)
	return nil
}

// newFrontendAssetHandler creates the guarded catch-all handler. It runs after
// host and source-plugin routes, so concrete plugin routes get first chance.
func newFrontendAssetHandler(
	subFS fs.FS,
	pluginSvc pluginsvc.Service,
	devProxy http.Handler,
	workspaceBasePath string,
) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		requestPath := normalizeRequestPath(r.URL.Path)
		if serveRuntimePluginAsset(r, pluginSvc, requestPath) {
			return
		}
		if isRootWorkspaceBasePath(workspaceBasePath) && isRootWorkspaceReservedRequest(requestPath) {
			r.Response.WriteStatus(http.StatusNotFound)
			r.ExitAll()
			return
		}
		workspacePath, ok := trimWorkspaceRequestPath(requestPath, workspaceBasePath)
		if !ok {
			r.Response.WriteStatus(http.StatusNotFound)
			r.ExitAll()
			return
		}
		if devProxy != nil {
			serveFrontendDevProxy(r, devProxy, requestPath, workspaceBasePath)
			r.ExitAll()
			return
		}
		if serveEmbeddedFrontendAsset(r, subFS, workspacePath) {
			return
		}
		serveSPAFallback(r, subFS)
	}
}

// serveFrontendDevProxy forwards workspace requests to Vite. Vite requires the
// configured base path to include its trailing slash, so exact `/admin` style
// requests are normalized before proxying.
func serveFrontendDevProxy(
	r *ghttp.Request,
	devProxy http.Handler,
	requestPath string,
	workspaceBasePath string,
) {
	proxyRequest := r.Request
	if strings.Trim(requestPath, "/") == strings.Trim(workspaceBasePath, "/") {
		proxyRequest = r.Request.Clone(r.Context())
		proxyRequest.URL.Path = workspaceBasePath
		if !strings.HasSuffix(proxyRequest.URL.Path, "/") {
			proxyRequest.URL.Path += "/"
		}
		proxyRequest.URL.RawPath = ""
	}
	devProxy.ServeHTTP(r.Response.RawWriter(), proxyRequest)
}

// newFrontendDevServerProxy builds the optional development reverse proxy used
// by linactl dev. Production leaves the env unset and serves embedded assets.
func newFrontendDevServerProxy() (http.Handler, error) {
	rawURL := strings.TrimSpace(os.Getenv(frontendDevServerURLEnv))
	if rawURL == "" {
		return nil, nil
	}
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if target.Scheme != "http" && target.Scheme != "https" {
		return nil, url.InvalidHostError("frontend dev server URL must use http or https")
	}
	if strings.TrimSpace(target.Host) == "" {
		return nil, url.InvalidHostError("frontend dev server URL must include host")
	}
	return httputil.NewSingleHostReverseProxy(target), nil
}

// serveRuntimePluginAsset serves versioned dynamic plugin frontend assets when
// the request path belongs to the public plugin-asset namespace.
func serveRuntimePluginAsset(
	r *ghttp.Request,
	pluginSvc pluginsvc.Service,
	path string,
) bool {
	// Plugin public assets must be checked before the host falls back to the
	// embedded frontend bundle. They are governed by plugin ID, version,
	// public_assets declarations, enabled state, and tenant availability.
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
	assetPath string,
) bool {
	content, err := fs.ReadFile(subFS, assetPath)
	if err != nil {
		return false
	}
	contentType := mime.TypeByExtension(path.Ext(assetPath))
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	r.Response.Header().Set("Content-Type", contentType)
	r.Response.Write(content)
	r.ExitAll()
	return true
}

// serveSPAFallback serves index.html directly for unmatched frontend routes so
// browser refreshes avoid net/http FileServer's index.html redirect behavior.
func serveSPAFallback(r *ghttp.Request, subFS fs.FS) {
	if !serveEmbeddedFrontendAsset(r, subFS, "index.html") {
		r.Response.WriteStatus(http.StatusNotFound)
	}
	r.ExitAll()
}

// normalizeRequestPath trims the leading slash while preserving sub-paths.
func normalizeRequestPath(rawPath string) string {
	return strings.TrimPrefix(strings.TrimSpace(rawPath), "/")
}

// normalizeWorkspaceRequestBasePath returns a rooted workspace base path for
// route registration and request-prefix checks.
func normalizeWorkspaceRequestBasePath(basePath string) string {
	normalized := strings.Trim(strings.TrimSpace(basePath), "/")
	if normalized == "" {
		return "/"
	}
	return "/" + normalized
}

// trimWorkspaceRequestPath removes the workspace base path from an incoming
// request and returns the embedded-asset path that should be served.
func trimWorkspaceRequestPath(requestPath string, workspaceBasePath string) (string, bool) {
	normalizedRequestPath := strings.Trim(requestPath, "/")
	if isRootWorkspaceBasePath(workspaceBasePath) {
		if normalizedRequestPath == "" {
			return "index.html", true
		}
		return normalizedRequestPath, true
	}
	if normalizedRequestPath == "" {
		return "", false
	}
	normalizedWorkspaceBasePath := strings.Trim(workspaceBasePath, "/")
	if normalizedRequestPath != normalizedWorkspaceBasePath &&
		!strings.HasPrefix(normalizedRequestPath, normalizedWorkspaceBasePath+"/") {
		return "", false
	}
	assetPath := strings.TrimPrefix(normalizedRequestPath, normalizedWorkspaceBasePath)
	assetPath = strings.Trim(assetPath, "/")
	if assetPath == "" {
		return "index.html", true
	}
	return assetPath, true
}

// isRootWorkspaceBasePath reports whether the admin workspace intentionally
// owns the public root route for dedicated-domain deployments.
func isRootWorkspaceBasePath(workspaceBasePath string) bool {
	return strings.Trim(strings.TrimSpace(workspaceBasePath), "/") == ""
}

// isRootWorkspaceReservedRequest keeps root-mounted workspace fallback from
// swallowing host APIs, plugin APIs, hosted plugin assets, or the OpenAPI JSON
// endpoint. Concrete routes are registered earlier; this guard covers misses.
func isRootWorkspaceReservedRequest(requestPath string) bool {
	normalizedPath := strings.Trim(strings.TrimSpace(requestPath), "/")
	if normalizedPath == "" {
		return false
	}
	for _, reserved := range []string{"api", pluginhost.PluginAPINamespaceSegment, pluginhost.HostedAssetPathSegment} {
		if normalizedPath == reserved || strings.HasPrefix(normalizedPath, reserved+"/") {
			return true
		}
	}
	return normalizedPath == "api.json"
}

// parsePluginAssetRequestPath splits one public `/x-assets/...` request
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
	if len(pathParts) < 3 || pathParts[0] != pluginhost.HostedAssetPathSegment {
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
