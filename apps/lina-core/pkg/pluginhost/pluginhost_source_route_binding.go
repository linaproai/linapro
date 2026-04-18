// This file defines the source-plugin route binding snapshot captured by the
// host during route registration for later governance and OpenAPI projection.

package pluginhost

import (
	"context"
	"reflect"
	"strings"

	"github.com/gogf/gf/v2/util/gmeta"
	"github.com/gogf/gf/v2/util/gtag"
)

// routeMethodAll is the normalized wildcard method marker used by route capture.
const routeMethodAll = "ALL"

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
	return normalizeRouteMethod(b.Method) + " " + normalizeRoutePattern(b.Path)
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

// captureRouteBindings captures one or more route bindings from either a
// function handler or a controller object.
func captureRouteBindings(
	pluginID string,
	prefix string,
	explicitPattern string,
	explicitMethod string,
	handler interface{},
) []SourceRouteBinding {
	reflectType := reflect.TypeOf(handler)
	if reflectType == nil {
		return nil
	}
	switch reflectType.Kind() {
	case reflect.Func:
		return captureFunctionRouteBindings(pluginID, prefix, explicitPattern, explicitMethod, handler)
	case reflect.Struct, reflect.Pointer:
		return captureObjectRouteBindings(pluginID, prefix, explicitPattern, handler)
	default:
		return nil
	}
}

// captureFunctionRouteBindings captures the effective methods and path for one
// function handler, consulting GoFrame request metadata when available.
func captureFunctionRouteBindings(
	pluginID string,
	prefix string,
	explicitPattern string,
	explicitMethod string,
	handler interface{},
) []SourceRouteBinding {
	var (
		routePath    = normalizeRoutePattern(explicitPattern)
		methods      = expandRouteMethods(explicitMethod)
		documentable = isDocumentableRouteHandler(handler)
	)
	if documentable {
		if reqMetaPath, ok := readHandlerMetaPath(handler); ok {
			routePath = normalizeRoutePattern(reqMetaPath)
		}
		if reqMetaMethods, ok := readHandlerMetaMethods(handler); ok {
			methods = reqMetaMethods
		}
	}
	if len(methods) == 0 {
		methods = []string{routeMethodAll}
	}

	finalPath := joinRoutePatterns(prefix, routePath)
	bindings := make([]SourceRouteBinding, 0, len(methods))
	for _, method := range methods {
		bindings = append(bindings, SourceRouteBinding{
			PluginID:     strings.TrimSpace(pluginID),
			Method:       normalizeRouteMethod(method),
			Path:         finalPath,
			Handler:      handler,
			Documentable: documentable,
		})
	}
	return bindings
}

// captureObjectRouteBindings expands one controller object into per-method route
// bindings using GoFrame's conventional method naming rules.
func captureObjectRouteBindings(
	pluginID string,
	prefix string,
	explicitPattern string,
	object interface{},
) []SourceRouteBinding {
	reflectValue := reflect.ValueOf(object)
	if !reflectValue.IsValid() {
		return nil
	}
	if reflectValue.Kind() == reflect.Struct {
		newValue := reflect.New(reflectValue.Type())
		newValue.Elem().Set(reflectValue)
		reflectValue = newValue
	}
	if reflectValue.Kind() != reflect.Pointer || reflectValue.Elem().Kind() != reflect.Struct {
		return nil
	}

	var (
		reflectType = reflectValue.Type()
		structName  = reflectType.Elem().Name()
		bindings    = make([]SourceRouteBinding, 0)
	)
	for i := 0; i < reflectValue.NumMethod(); i++ {
		methodName := reflectType.Method(i).Name
		if methodName == "Init" || methodName == "Shut" {
			continue
		}

		methodValue := reflectValue.Method(i)
		if !methodValue.IsValid() {
			continue
		}
		methodBindings := captureFunctionRouteBindings(
			pluginID,
			prefix,
			mergeRouteMethodPattern(explicitPattern, structName, methodName),
			routeMethodAll,
			methodValue.Interface(),
		)
		bindings = append(bindings, methodBindings...)
	}
	return bindings
}

// isDocumentableRouteHandler reports whether the handler matches the standard
// GoFrame `(ctx, *Req) (*Res, error)` shape required for OpenAPI projection.
func isDocumentableRouteHandler(handler interface{}) bool {
	reflectType := reflect.TypeOf(handler)
	if reflectType == nil || reflectType.Kind() != reflect.Func {
		return false
	}
	if reflectType.NumIn() != 2 || reflectType.NumOut() != 2 {
		return false
	}
	if !reflectType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return false
	}
	if !reflectType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return false
	}
	return reflectType.In(1).Kind() == reflect.Pointer && reflectType.In(1).Elem().Kind() == reflect.Struct
}

// readHandlerMetaPath reads the GoFrame `path` metadata from the handler
// request DTO when the handler is documentable.
func readHandlerMetaPath(handler interface{}) (string, bool) {
	reqObject, ok := newHandlerReqObject(handler)
	if !ok {
		return "", false
	}
	metaPath := strings.TrimSpace(gmeta.Get(reqObject, gtag.Path).String())
	if metaPath == "" {
		return "", false
	}
	return metaPath, true
}

// readHandlerMetaMethods reads the GoFrame `method` metadata from the handler
// request DTO when the handler is documentable.
func readHandlerMetaMethods(handler interface{}) ([]string, bool) {
	reqObject, ok := newHandlerReqObject(handler)
	if !ok {
		return nil, false
	}
	metaMethod := strings.TrimSpace(gmeta.Get(reqObject, gtag.Method).String())
	if metaMethod == "" {
		return nil, false
	}
	return expandRouteMethods(metaMethod), true
}

// newHandlerReqObject allocates a fresh request DTO instance for metadata
// inspection from a documentable handler.
func newHandlerReqObject(handler interface{}) (interface{}, bool) {
	reflectType := reflect.TypeOf(handler)
	if !isDocumentableRouteHandler(handler) || reflectType == nil {
		return nil, false
	}
	return reflect.New(reflectType.In(1).Elem()).Interface(), true
}

// expandRouteMethods splits a comma-separated method declaration into
// normalized method names.
func expandRouteMethods(method string) []string {
	trimmed := strings.TrimSpace(method)
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, ",")
	methods := make([]string, 0, len(parts))
	for _, part := range parts {
		normalized := normalizeRouteMethod(part)
		if normalized == "" {
			continue
		}
		methods = append(methods, normalized)
	}
	return methods
}

// normalizeRouteMethod canonicalizes an HTTP method and falls back to ALL when
// the input is empty.
func normalizeRouteMethod(method string) string {
	trimmed := strings.TrimSpace(method)
	if trimmed == "" {
		return routeMethodAll
	}
	normalized := strings.ToUpper(trimmed)
	if normalized == routeMethodAll {
		return routeMethodAll
	}
	return normalized
}

// normalizeRoutePattern canonicalizes one route path for stable capture keys
// and path composition.
func normalizeRoutePattern(pattern string) string {
	trimmed := strings.TrimSpace(pattern)
	if trimmed == "" || trimmed == "/" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/")
}

// joinRoutePatterns joins a group prefix and child pattern into one normalized
// public route path.
func joinRoutePatterns(prefix string, pattern string) string {
	normalizedPrefix := normalizeRoutePattern(prefix)
	normalizedPattern := normalizeRoutePattern(pattern)
	if normalizedPrefix == "/" {
		return normalizedPattern
	}
	if normalizedPattern == "/" {
		return normalizedPrefix
	}
	return normalizedPrefix + "/" + strings.TrimLeft(normalizedPattern, "/")
}

// mergeRouteMethodPattern resolves GoFrame controller placeholders and implicit
// method suffixes for controller-object route capture.
func mergeRouteMethodPattern(pattern string, structName string, methodName string) string {
	normalizedPattern := normalizeRoutePattern(pattern)
	structSegment := normalizeRouteName(structName)
	methodSegment := normalizeRouteName(methodName)

	merged := strings.ReplaceAll(normalizedPattern, "{.struct}", structSegment)
	if strings.Contains(merged, "{.method}") {
		return strings.ReplaceAll(merged, "{.method}", methodSegment)
	}
	if normalizedPattern == "/" {
		return "/" + methodSegment
	}
	return strings.TrimRight(merged, "/") + "/" + methodSegment
}

// normalizeRouteName converts a Go-style exported identifier to the kebab-case
// route segment GoFrame uses for strict routes.
func normalizeRouteName(name string) string {
	if name == "" {
		return ""
	}
	var builder strings.Builder
	for index, runeValue := range name {
		if index > 0 && runeValue >= 'A' && runeValue <= 'Z' {
			builder.WriteByte('-')
		}
		builder.WriteString(strings.ToLower(string(runeValue)))
	}
	return builder.String()
}
