// This file defines the source-plugin visible API-documentation contract and
// route operation-key helpers.

package apidoccap

import (
	"context"
	"regexp"
	"strings"
)

// RouteTextInput defines one API-documentation route text lookup request.
type RouteTextInput struct {
	// OperationKey is the stable apidoc operation key base when known.
	OperationKey string
	// Method is the HTTP method used when OperationKey must be derived from Path.
	Method string
	// Path is the normalized public route path used when OperationKey is empty.
	Path string
	// FallbackTitle is returned when the apidoc catalog has no tag translation.
	FallbackTitle string
	// FallbackSummary is returned when the apidoc catalog has no summary translation.
	FallbackSummary string
}

// RouteTextOutput contains localized route text for one audit-log record.
type RouteTextOutput struct {
	// Title is the localized module tag.
	Title string
	// Summary is the localized operation summary.
	Summary string
}

// Service defines the API-documentation operations published to source plugins.
type Service interface {
	// ResolveRouteText resolves one route's localized module tag and operation summary.
	ResolveRouteText(ctx context.Context, input RouteTextInput) RouteTextOutput
	// ResolveRouteTexts resolves multiple route texts with one apidoc catalog load.
	ResolveRouteTexts(ctx context.Context, inputs []RouteTextInput) []RouteTextOutput
	// FindRouteTitleOperationKeys finds localized module tag operation keys by keyword.
	FindRouteTitleOperationKeys(ctx context.Context, keyword string) []string
}

// BuildOperationKeyFromPath returns the path-derived apidoc operation key base
// used for dynamic plugin routes and non-DTO fallback routes.
func BuildOperationKeyFromPath(path string, method string) string {
	normalizedPath := normalizeOpenAPIPath(path)
	if normalizedPath == "" {
		return ""
	}
	return buildOpenAPIPathOperationKey(normalizedPath, strings.ToLower(strings.TrimSpace(method)))
}

// BuildDynamicOperationKey returns the path-derived apidoc operation key base
// for one dynamic-plugin route.
func BuildDynamicOperationKey(path string, method string) string {
	return BuildOperationKeyFromPath(path, method)
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

// buildOpenAPIPathOperationKey returns a stable structural key base for one operation.
func buildOpenAPIPathOperationKey(pathName string, method string) string {
	segments := openAPIPathSegments(pathName)
	if isDynamicPluginOpenAPIPath(pathName) {
		pluginID, remainingSegments := dynamicPluginOpenAPIPathParts(segments)
		remainingPath := dynamicPluginRouteKeyPath(remainingSegments)
		if remainingPath == "" {
			remainingPath = "root"
		}
		return "plugins." + pluginID + ".paths." + sanitizeOpenAPIKeyPart(method) + "." + remainingPath
	}
	return buildOpenAPIPathKey(pathName) + "." + sanitizeOpenAPIKeyPart(method)
}

// dynamicPluginRouteKeyPath returns a generic path-derived key fragment without
// interpreting plugin-owned route segments.
func dynamicPluginRouteKeyPath(segments []string) string {
	return strings.Join(segments, ".")
}

// buildOpenAPIPathKey returns a stable structural key base for a path item.
func buildOpenAPIPathKey(pathName string) string {
	return "core.paths." + sanitizeOpenAPIPathKey(pathName)
}

// isDynamicPluginOpenAPIPath reports whether a public OpenAPI path belongs to the dynamic-plugin namespace.
func isDynamicPluginOpenAPIPath(pathName string) bool {
	segments := openAPIPathSegments(pathName)
	_, _, ok := dynamicPluginOpenAPIPath(segments)
	return ok
}

// dynamicPluginOpenAPIPathParts returns the stable plugin key segment and the
// plugin-owned route path segments for dynamic routes.
func dynamicPluginOpenAPIPathParts(segments []string) (string, []string) {
	pluginIndex, routeStart, ok := dynamicPluginOpenAPIPath(segments)
	if !ok {
		return "", nil
	}
	return sanitizeOpenAPIKeyPart(segments[pluginIndex]), segments[routeStart:]
}

// dynamicPluginOpenAPIPath detects `/x/{pluginId}/...` paths after OpenAPI key sanitization.
func dynamicPluginOpenAPIPath(segments []string) (pluginIndex int, routeStart int, ok bool) {
	if len(segments) >= 2 && segments[0] == "x" {
		return 1, 2, true
	}
	return 0, 0, false
}

// sanitizeOpenAPIPathKey converts an OpenAPI path into dot-separated key parts.
func sanitizeOpenAPIPathKey(pathName string) string {
	segments := openAPIPathSegments(pathName)
	if len(segments) == 0 {
		return "root"
	}
	return strings.Join(segments, ".")
}

// openAPIPathSegments normalizes path segments for use in translation keys.
func openAPIPathSegments(pathName string) []string {
	var segments []string
	for _, segment := range strings.Split(strings.Trim(pathName, "/"), "/") {
		if strings.TrimSpace(segment) == "" {
			continue
		}
		segments = append(segments, sanitizeOpenAPIKeyPart(segment))
	}
	return segments
}

// openAPIKeyInvalidCharsPattern matches characters that cannot be used in apidoc keys.
var openAPIKeyInvalidCharsPattern = regexp.MustCompile(`[^A-Za-z0-9_]+`)

// sanitizeOpenAPIKeyPart normalizes one key segment for safe JSON-object keys.
func sanitizeOpenAPIKeyPart(part string) string {
	trimmedPart := strings.TrimSpace(part)
	trimmedPart = strings.Trim(trimmedPart, "{}")
	trimmedPart = strings.ReplaceAll(trimmedPart, "-", "_")
	trimmedPart = openAPIKeyInvalidCharsPattern.ReplaceAllString(trimmedPart, "_")
	trimmedPart = strings.Trim(trimmedPart, "_")
	if trimmedPart == "" {
		return "empty"
	}
	return trimmedPart
}
