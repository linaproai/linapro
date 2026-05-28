// This file loads the default admin workspace entry path and validates that it
// cannot overlap host APIs, plugin APIs, or plugin assets.

package config

import (
	"context"
	"path"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// defaultWorkspaceBasePath is the default browser entry for the built-in admin
// workspace. It is intentionally not "/" so source plugins can own root routes.
const defaultWorkspaceBasePath = "/admin"

// defaultWorkspaceRouterMode is the default vue-router history backend used
// by the admin workspace. Hash mode is the default because it works without
// server-side SPA fallback configuration, matching the bundled deployment.
const defaultWorkspaceRouterMode = "hash"

// WorkspaceRouterModeHash represents the hash-based vue-router history backend.
const WorkspaceRouterModeHash = "hash"

// WorkspaceRouterModeHistory represents the HTML5 history vue-router backend.
const WorkspaceRouterModeHistory = "history"

// WorkspaceConfig holds static admin workspace routing settings.
type WorkspaceConfig struct {
	// BasePath is the admin workspace entry path.
	BasePath string `json:"basePath"`
	// RouterMode selects the vue-router history backend used by the
	// workspace. Allowed values are "hash" (default, no server SPA fallback
	// required) and "history" (requires the host or a reverse proxy to
	// serve the SPA index for unknown paths under BasePath). The backend
	// uses this together with BasePath to compose full workspace URLs for
	// OAuth handoff and other post-login redirects so backend-issued URLs
	// stay in sync with the frontend bundle's actual routing.
	RouterMode string `json:"routerMode"`
}

// getStaticWorkspaceConfig lazily loads and validates workspace routing
// settings from config.yaml because the workspace entry is startup-scoped.
func (s *serviceImpl) getStaticWorkspaceConfig(ctx context.Context) *WorkspaceConfig {
	return processStaticConfigCaches.workspace.load(func() *WorkspaceConfig {
		cfg := &WorkspaceConfig{
			BasePath:   defaultWorkspaceBasePath,
			RouterMode: defaultWorkspaceRouterMode,
		}
		mustScanConfig(ctx, "workspace", cfg)
		cfg.BasePath = mustNormalizeWorkspaceBasePath(cfg.BasePath)
		cfg.RouterMode = mustNormalizeWorkspaceRouterMode(cfg.RouterMode)
		return cfg
	})
}

// GetWorkspace reads the admin workspace routing config from config.yaml.
func (s *serviceImpl) GetWorkspace(ctx context.Context) *WorkspaceConfig {
	return cloneWorkspaceConfig(s.getStaticWorkspaceConfig(ctx))
}

// GetWorkspaceBasePath returns the normalized admin workspace entry path.
func (s *serviceImpl) GetWorkspaceBasePath(ctx context.Context) string {
	cfg := s.getStaticWorkspaceConfig(ctx)
	if cfg == nil || strings.TrimSpace(cfg.BasePath) == "" {
		return defaultWorkspaceBasePath
	}
	return cfg.BasePath
}

// GetWorkspaceRouterMode returns the normalized vue-router history backend
// (hash or history) used by the admin workspace.
func (s *serviceImpl) GetWorkspaceRouterMode(ctx context.Context) string {
	cfg := s.getStaticWorkspaceConfig(ctx)
	if cfg == nil || strings.TrimSpace(cfg.RouterMode) == "" {
		return defaultWorkspaceRouterMode
	}
	return cfg.RouterMode
}

// BuildWorkspaceRouteURL composes the absolute URL path that the browser
// must visit to land on one workspace route. It always uses the configured
// base path and inserts the vue-router hash separator when the workspace is
// built with hash history, so backend-issued redirects (such as OAuth
// handoff) always reach the SPA regardless of which router backend the
// frontend uses.
func (s *serviceImpl) BuildWorkspaceRouteURL(ctx context.Context, routePath string, rawQuery string) string {
	return composeWorkspaceRouteURL(s.GetWorkspaceBasePath(ctx), s.GetWorkspaceRouterMode(ctx), routePath, rawQuery)
}

// composeWorkspaceRouteURL assembles one workspace URL from raw inputs
// without reading from the config store, so unit tests can exercise the
// routing math without standing up the runtime config service.
func composeWorkspaceRouteURL(basePath string, routerMode string, routePath string, rawQuery string) string {
	normalizedBase := strings.TrimSpace(basePath)
	if normalizedBase == "" {
		normalizedBase = defaultWorkspaceBasePath
	}
	normalizedBase = strings.TrimRight(normalizedBase, "/")
	normalizedRoute := strings.TrimSpace(routePath)
	if normalizedRoute == "" {
		normalizedRoute = "/"
	}
	if !strings.HasPrefix(normalizedRoute, "/") {
		normalizedRoute = "/" + normalizedRoute
	}
	normalizedQuery := strings.TrimSpace(rawQuery)
	normalizedQuery = strings.TrimPrefix(normalizedQuery, "?")
	normalizedMode := strings.TrimSpace(routerMode)
	if normalizedMode == "" {
		normalizedMode = defaultWorkspaceRouterMode
	}
	var builder strings.Builder
	if normalizedBase != "" && normalizedBase != "/" {
		builder.WriteString(normalizedBase)
	}
	if normalizedMode == WorkspaceRouterModeHash {
		builder.WriteString("/#")
		builder.WriteString(normalizedRoute)
	} else {
		builder.WriteString(normalizedRoute)
	}
	if normalizedQuery != "" {
		builder.WriteString("?")
		builder.WriteString(normalizedQuery)
	}
	return builder.String()
}

// mustNormalizeWorkspaceRouterMode validates a router-mode value and
// returns the canonical form. Invalid values fail fast because
// backend-issued URLs must not silently fall back to a different routing
// scheme than the bundled frontend.
func mustNormalizeWorkspaceRouterMode(raw string) string {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "" {
		return defaultWorkspaceRouterMode
	}
	switch trimmed {
	case WorkspaceRouterModeHash, WorkspaceRouterModeHistory:
		return trimmed
	default:
		panic(workspaceStartupDiagnosticError(
			"workspace.routerMode",
			"must be one of: hash, history",
			"set workspace.routerMode=hash to match the bundled frontend",
		))
	}
}

// mustNormalizeWorkspaceBasePath normalizes and validates the workspace base
// path. Invalid paths fail fast because route binding must not continue with an
// ambiguous catch-all frontend fallback.
func mustNormalizeWorkspaceBasePath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		panic(workspaceStartupDiagnosticError(
			"workspace.basePath", "required", "set workspace.basePath=/admin",
		))
	}
	if strings.Contains(trimmed, "*") {
		panic(workspaceStartupDiagnosticError(
			"workspace.basePath",
			"wildcards are not allowed",
			"use a concrete path such as /admin or /",
		))
	}
	if strings.Contains(trimmed, "?") || strings.Contains(trimmed, "#") {
		panic(workspaceStartupDiagnosticError(
			"workspace.basePath",
			"query strings and fragments are not allowed",
			"set only the URL path segment, for example /admin",
		))
	}
	if strings.Contains(trimmed, "://") || !strings.HasPrefix(trimmed, "/") {
		panic(workspaceStartupDiagnosticError(
			"workspace.basePath",
			"must be an absolute URL path",
			"prefix the path with /, for example /admin",
		))
	}
	normalized := path.Clean(trimmed)
	if normalized == "." {
		panic(workspaceStartupDiagnosticError(
			"workspace.basePath",
			"must be an absolute URL path",
			"set /admin or /",
		))
	}
	if strings.Contains(normalized, "//") {
		panic(workspaceStartupDiagnosticError(
			"workspace.basePath",
			"must not contain empty path segments",
			"remove duplicate slashes",
		))
	}
	for _, reserved := range workspaceReservedBasePathPrefixes() {
		if normalized == reserved || strings.HasPrefix(normalized, reserved+"/") {
			panic(workspaceStartupDiagnosticError(
				"workspace.basePath",
				"conflicts with reserved namespace "+reserved,
				"use /admin or another non-reserved path",
			))
		}
	}
	return normalized
}

// workspaceReservedBasePathPrefixes lists host-owned public route namespaces
// that the admin workspace fallback must not occupy.
func workspaceReservedBasePathPrefixes() []string {
	return []string{
		"/api",
		"/api/v1",
		"/x",
		"/x-assets",
		"/plugin-assets",
	}
}

// workspaceStartupDiagnosticError formats static workspace configuration
// failures with the broken field and a concrete remediation.
func workspaceStartupDiagnosticError(field string, reason string, fix string) error {
	return gerror.Newf("workspace startup diagnostic field=%s reason=%s fix=%s", field, reason, fix)
}
