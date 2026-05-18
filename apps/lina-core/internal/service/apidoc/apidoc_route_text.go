// This file exposes route-level text resolution helpers so runtime audit logs
// can reuse the same apidoc i18n resources as the OpenAPI document.

package apidoc

import (
	"context"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

const (
	// routeTextTagsSuffix is the metadata key suffix used for operation module tags.
	routeTextTagsSuffix = ".meta.tags"
	// routeTextSummarySuffix is the metadata key suffix used for operation summaries.
	routeTextSummarySuffix = ".meta.summary"
)

// RouteTextInput defines one apidoc route text lookup request.
type RouteTextInput struct {
	// OperationKey is the stable apidoc operation key base when known.
	OperationKey string
	// Method is the HTTP method used when OperationKey must be derived from Path.
	Method string
	// Path is the normalized public route path used when OperationKey is empty.
	Path string
	// FallbackTitle is returned when the current apidoc catalog has no tag translation.
	FallbackTitle string
	// FallbackSummary is returned when the current apidoc catalog has no summary translation.
	FallbackSummary string
}

// RouteTextOutput contains localized route text for one audit-log record.
type RouteTextOutput struct {
	// Title is the localized module tag.
	Title string
	// Summary is the localized operation summary.
	Summary string
}

// ResolveRouteText resolves one route's localized module tag and operation
// summary from the dedicated apidoc i18n catalog.
func (s *serviceImpl) ResolveRouteText(ctx context.Context, input RouteTextInput) RouteTextOutput {
	if s == nil || s.i18nSvc == nil {
		return routeTextFallback(input)
	}

	localizer := &openAPILocalizer{
		catalog: s.loadOpenAPIMessageCatalog(ctx, s.i18nSvc.GetLocale(ctx)),
	}
	return resolveRouteTextWithLocalizer(localizer, input)
}

// ResolveRouteTexts resolves multiple route text projections with one apidoc
// catalog load so operation-log pages do not reload plugin metadata per row.
func (s *serviceImpl) ResolveRouteTexts(ctx context.Context, inputs []RouteTextInput) []RouteTextOutput {
	outputs := make([]RouteTextOutput, 0, len(inputs))
	if len(inputs) == 0 {
		return outputs
	}
	if s == nil || s.i18nSvc == nil {
		for _, input := range inputs {
			outputs = append(outputs, routeTextFallback(input))
		}
		return outputs
	}

	localizer := &openAPILocalizer{
		catalog: s.loadOpenAPIMessageCatalog(ctx, s.i18nSvc.GetLocale(ctx)),
	}
	for _, input := range inputs {
		outputs = append(outputs, resolveRouteTextWithLocalizer(localizer, input))
	}
	return outputs
}

// FindRouteTitleOperationKeys finds operation key bases whose localized module
// tag contains the given keyword in the current request locale.
func (s *serviceImpl) FindRouteTitleOperationKeys(ctx context.Context, keyword string) []string {
	if s == nil || s.i18nSvc == nil || strings.TrimSpace(keyword) == "" {
		return []string{}
	}

	normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
	catalog := s.loadOpenAPIMessageCatalog(ctx, s.i18nSvc.GetLocale(ctx))
	matches := make(map[string]struct{})
	for key, value := range catalog {
		if !strings.Contains(strings.ToLower(value), normalizedKeyword) {
			continue
		}
		if operationKey := trimRouteTitleMetadataSuffix(key); operationKey != "" {
			matches[operationKey] = struct{}{}
		}
	}

	keys := make([]string, 0, len(matches))
	for key := range matches {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// routeTextFallback returns untranslated route text fallback values.
func routeTextFallback(input RouteTextInput) RouteTextOutput {
	return RouteTextOutput{
		Title:   input.FallbackTitle,
		Summary: input.FallbackSummary,
	}
}

// resolveRouteTextWithLocalizer resolves one route text from an already loaded
// apidoc catalog.
func resolveRouteTextWithLocalizer(localizer *openAPILocalizer, input RouteTextInput) RouteTextOutput {
	output := routeTextFallback(input)
	if localizer == nil {
		return output
	}
	keyBase := strings.TrimSpace(input.OperationKey)
	if keyBase == "" {
		keyBase = BuildRouteOperationKeyFromPath(input.Path, input.Method)
	}
	if keyBase == "" {
		return output
	}
	output.Title = localizer.translate(keyBase+routeTextTagsSuffix, input.FallbackTitle)
	output.Summary = localizer.translate(keyBase+routeTextSummarySuffix, input.FallbackSummary)
	return output
}

// BuildRouteOperationKeyFromHandlerType returns the static-route apidoc key
// base for a GoFrame strict handler function type.
func BuildRouteOperationKeyFromHandlerType(handlerType reflect.Type) string {
	if handlerType == nil || handlerType.Kind() != reflect.Func || handlerType.NumIn() != 2 {
		return ""
	}
	return BuildRouteOperationKeyFromRequestType(handlerType.In(1))
}

// BuildRouteOperationKeyFromRequestType returns the static-route apidoc key
// base for a GoFrame request DTO type.
func BuildRouteOperationKeyFromRequestType(reqType reflect.Type) string {
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

// BuildRouteOperationKeyFromPath returns the path-derived apidoc operation key
// base used for dynamic plugin routes and non-DTO fallback routes.
func BuildRouteOperationKeyFromPath(path string, method string) string {
	normalizedPath := normalizeOpenAPIPath(path)
	if normalizedPath == "" {
		return ""
	}
	return buildOpenAPIPathOperationKey(normalizedPath, strings.ToLower(strings.TrimSpace(method)))
}

// trimRouteTitleMetadataSuffix strips `.meta.tags` or indexed tag suffixes from
// an apidoc catalog key and returns the operation key base.
func trimRouteTitleMetadataSuffix(key string) string {
	trimmedKey := strings.TrimSpace(key)
	if strings.HasSuffix(trimmedKey, routeTextTagsSuffix) {
		return strings.TrimSuffix(trimmedKey, routeTextTagsSuffix)
	}

	index := strings.LastIndex(trimmedKey, routeTextTagsSuffix+".")
	if index < 0 {
		return ""
	}
	indexText := trimmedKey[index+len(routeTextTagsSuffix)+1:]
	if _, err := strconv.Atoi(indexText); err != nil {
		return ""
	}
	return trimmedKey[:index]
}
