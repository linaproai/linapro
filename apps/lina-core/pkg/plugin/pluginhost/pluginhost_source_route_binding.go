// This file defines the source-plugin route binding snapshot captured by the
// host during route registration for later governance and OpenAPI projection.

package pluginhost

import "strings"

// sourceRouteMethodAll is the normalized wildcard method marker used by route
// binding keys.
const sourceRouteMethodAll = "ALL"

// SourceRouteBinding stores one plugin-owned route binding captured during host
// route registration.
type SourceRouteBinding struct {
	// PluginID is the owning source-plugin identifier.
	PluginID string
	// Method is the resolved HTTP method, such as GET or POST.
	Method string
	// Path is the resolved public route path registered on the host server.
	Path string
	// Handler is the bound route handler or bound object method.
	Handler interface{}
	// Documentable reports whether the handler uses GoFrame standard DTO routing
	// and can therefore be projected into the host OpenAPI document.
	Documentable bool
}

// Key returns the normalized uniqueness key of the binding.
func (b SourceRouteBinding) Key() string {
	return normalizeSourceRouteMethod(b.Method) + " " + normalizeSourceRoutePattern(b.Path)
}

// CloneSourceRouteBindings returns one detached copy of the given bindings.
func CloneSourceRouteBindings(bindings []SourceRouteBinding) []SourceRouteBinding {
	if len(bindings) == 0 {
		return []SourceRouteBinding{}
	}
	items := make([]SourceRouteBinding, len(bindings))
	copy(items, bindings)
	return items
}

// normalizeSourceRouteMethod canonicalizes an HTTP method and falls back to ALL when
// the input is empty.
func normalizeSourceRouteMethod(method string) string {
	trimmed := strings.TrimSpace(method)
	if trimmed == "" {
		return sourceRouteMethodAll
	}
	normalized := strings.ToUpper(trimmed)
	if normalized == sourceRouteMethodAll {
		return sourceRouteMethodAll
	}
	return normalized
}

// normalizeSourceRoutePattern canonicalizes one route path for stable capture keys.
func normalizeSourceRoutePattern(pattern string) string {
	trimmed := strings.TrimSpace(pattern)
	if trimmed == "" || trimmed == "/" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/")
}
