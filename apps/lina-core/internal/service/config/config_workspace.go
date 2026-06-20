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

// WorkspaceConfig holds static admin workspace routing settings.
type WorkspaceConfig struct {
	BasePath string `json:"basePath"` // BasePath is the admin workspace entry path.
}

// getStaticWorkspaceConfig lazily loads and validates workspace routing
// settings from config.yaml because the workspace entry is startup-scoped.
func (s *serviceImpl) getStaticWorkspaceConfig(ctx context.Context) *WorkspaceConfig {
	return processStaticConfigCaches.workspace.load(func() *WorkspaceConfig {
		cfg := &WorkspaceConfig{
			BasePath: defaultWorkspaceBasePath,
		}
		mustScanConfig(ctx, "workspace", cfg)
		cfg.BasePath = mustNormalizeWorkspaceBasePath(cfg.BasePath)
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
	}
}

// workspaceStartupDiagnosticError formats static workspace configuration
// failures with the broken field and a concrete remediation.
func workspaceStartupDiagnosticError(field string, reason string, fix string) error {
	return gerror.Newf("workspace startup diagnostic field=%s reason=%s fix=%s", field, reason, fix)
}
