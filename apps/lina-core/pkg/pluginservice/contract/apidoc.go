// This file defines the source-plugin visible API-documentation contract and
// route operation-key helpers.

package contract

import (
	"context"
	"reflect"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
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

// APIDocService defines the API-documentation operations published to source plugins.
type APIDocService interface {
	// ResolveRouteText resolves one route's localized module tag and operation summary.
	ResolveRouteText(ctx context.Context, input RouteTextInput) RouteTextOutput
	// ResolveRouteTexts resolves multiple route texts with one apidoc catalog load.
	ResolveRouteTexts(ctx context.Context, inputs []RouteTextInput) []RouteTextOutput
	// FindRouteTitleOperationKeys finds localized module tag operation keys by keyword.
	FindRouteTitleOperationKeys(ctx context.Context, keyword string) []string
}

// BuildOperationKeyFromHandler returns the apidoc operation key base for one
// GoFrame strict route handler.
func BuildOperationKeyFromHandler(handler *ghttp.HandlerItemParsed) string {
	if handler == nil || handler.Handler == nil {
		return ""
	}
	return buildOperationKeyFromHandlerType(handler.Handler.Info.Type)
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

// buildOperationKeyFromHandlerType returns the static-route apidoc key base for a handler function type.
func buildOperationKeyFromHandlerType(handlerType reflect.Type) string {
	if handlerType == nil || handlerType.Kind() != reflect.Func || handlerType.NumIn() != 2 {
		return ""
	}
	return buildOperationKeyFromRequestType(handlerType.In(1))
}

// buildOperationKeyFromRequestType returns the static-route apidoc key base for a request DTO type.
func buildOperationKeyFromRequestType(reqType reflect.Type) string {
	if reqType == nil {
		return ""
	}
	if reqType.Kind() == reflect.Pointer {
		reqType = reqType.Elem()
	}
	if reqType.Kind() != reflect.Struct || strings.TrimSpace(reqType.Name()) == "" {
		return ""
	}
	componentName := strings.ReplaceAll(reqType.PkgPath(), "/", ".") + "." + reqType.Name()
	return normalizeOpenAPIComponentKey(componentName)
}

// normalizeOpenAPIComponentKey converts GoFrame schema component names into stable apidoc keys.
func normalizeOpenAPIComponentKey(name string) string {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return ""
	}
	if strings.HasPrefix(trimmedName, "lina-plugin-") {
		return normalizeSourcePluginOpenAPIComponentKey(trimmedName)
	}
	replacements := []struct {
		old string
		new string
	}{
		{old: "lina-core.", new: "core."},
		{old: "lina-plugins.", new: "plugins."},
	}
	for _, replacement := range replacements {
		if strings.HasPrefix(trimmedName, replacement.old) {
			trimmedName = replacement.new + strings.TrimPrefix(trimmedName, replacement.old)
			break
		}
	}
	return sanitizeOpenAPIKey(trimmedName)
}

// normalizeSourcePluginOpenAPIComponentKey converts source-plugin schema component names into plugin keys.
func normalizeSourcePluginOpenAPIComponentKey(name string) string {
	trimmedName := strings.TrimSpace(name)
	withoutPrefix := strings.TrimPrefix(trimmedName, "lina-plugin-")
	pluginPart, rest, ok := strings.Cut(withoutPrefix, ".")
	if !ok || strings.TrimSpace(pluginPart) == "" {
		return sanitizeOpenAPIKey(trimmedName)
	}
	if strings.HasPrefix(rest, "backend.api.") {
		rest = strings.TrimPrefix(rest, "backend.")
	}
	return "plugins." + sanitizeOpenAPIKeyPart(pluginPart) + "." + sanitizeOpenAPIKey(rest)
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
		pluginID := sanitizeOpenAPIKeyPart(segments[3])
		remainingPath := strings.Join(segments[4:], ".")
		if remainingPath == "" {
			remainingPath = "root"
		}
		return "plugins." + pluginID + ".paths." + sanitizeOpenAPIKeyPart(method) + "." + remainingPath
	}
	return buildOpenAPIPathKey(pathName) + "." + sanitizeOpenAPIKeyPart(method)
}

// buildOpenAPIPathKey returns a stable structural key base for a path item.
func buildOpenAPIPathKey(pathName string) string {
	return "core.paths." + sanitizeOpenAPIPathKey(pathName)
}

// isDynamicPluginOpenAPIPath reports whether a public OpenAPI path belongs to the extension namespace.
func isDynamicPluginOpenAPIPath(pathName string) bool {
	segments := openAPIPathSegments(pathName)
	return len(segments) >= 4 && segments[0] == "api" && segments[1] == "v1" && segments[2] == "extensions"
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

// sanitizeOpenAPIKey normalizes a full key while preserving dots as hierarchy separators.
func sanitizeOpenAPIKey(key string) string {
	parts := strings.Split(strings.TrimSpace(key), ".")
	for index, part := range parts {
		parts[index] = sanitizeOpenAPIKeyPart(part)
	}
	return strings.Join(parts, ".")
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
