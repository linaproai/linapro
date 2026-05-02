// Package apidoc exposes a narrowed host API-documentation i18n contract to
// source plugins without requiring them to import host-internal service packages.
package apidoc

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	internalapidoc "lina-core/internal/service/apidoc"
	internalplugin "lina-core/internal/service/plugin"
)

// RouteTextInput defines one route text lookup request.
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

// Service defines the apidoc i18n operations published to source plugins.
type Service interface {
	// ResolveRouteText resolves one route's localized module tag and operation
	// summary from the dedicated apidoc i18n catalog.
	ResolveRouteText(ctx context.Context, input RouteTextInput) RouteTextOutput
	// ResolveRouteTexts resolves multiple route texts with one apidoc catalog load.
	ResolveRouteTexts(ctx context.Context, inputs []RouteTextInput) []RouteTextOutput
	// FindRouteTitleOperationKeys finds operation key bases whose localized
	// module tag contains the given keyword in the current request locale.
	FindRouteTitleOperationKeys(ctx context.Context, keyword string) []string
}

// serviceAdapter bridges the internal apidoc service into the published plugin contract.
type serviceAdapter struct {
	service internalapidoc.Service
}

// New creates and returns the published apidoc service adapter.
func New() Service {
	return &serviceAdapter{service: internalapidoc.New(nil, internalplugin.New(nil))}
}

// ResolveRouteText resolves one route's localized module tag and operation
// summary from the dedicated apidoc i18n catalog.
func (s *serviceAdapter) ResolveRouteText(ctx context.Context, input RouteTextInput) RouteTextOutput {
	if s == nil || s.service == nil {
		return RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary}
	}
	output := s.service.ResolveRouteText(ctx, internalapidoc.RouteTextInput{
		OperationKey:    input.OperationKey,
		Method:          input.Method,
		Path:            input.Path,
		FallbackTitle:   input.FallbackTitle,
		FallbackSummary: input.FallbackSummary,
	})
	return RouteTextOutput{Title: output.Title, Summary: output.Summary}
}

// ResolveRouteTexts resolves multiple route texts with one apidoc catalog load.
func (s *serviceAdapter) ResolveRouteTexts(ctx context.Context, inputs []RouteTextInput) []RouteTextOutput {
	outputs := make([]RouteTextOutput, 0, len(inputs))
	if s == nil || s.service == nil {
		for _, input := range inputs {
			outputs = append(outputs, RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary})
		}
		return outputs
	}
	internalInputs := make([]internalapidoc.RouteTextInput, 0, len(inputs))
	for _, input := range inputs {
		internalInputs = append(internalInputs, internalapidoc.RouteTextInput{
			OperationKey:    input.OperationKey,
			Method:          input.Method,
			Path:            input.Path,
			FallbackTitle:   input.FallbackTitle,
			FallbackSummary: input.FallbackSummary,
		})
	}
	for _, output := range s.service.ResolveRouteTexts(ctx, internalInputs) {
		outputs = append(outputs, RouteTextOutput{Title: output.Title, Summary: output.Summary})
	}
	return outputs
}

// FindRouteTitleOperationKeys finds operation key bases whose localized module
// tag contains the given keyword in the current request locale.
func (s *serviceAdapter) FindRouteTitleOperationKeys(ctx context.Context, keyword string) []string {
	if s == nil || s.service == nil {
		return []string{}
	}
	return s.service.FindRouteTitleOperationKeys(ctx, keyword)
}

// BuildOperationKeyFromHandler returns the apidoc operation key base for one
// GoFrame strict route handler.
func BuildOperationKeyFromHandler(handler *ghttp.HandlerItemParsed) string {
	if handler == nil || handler.Handler == nil {
		return ""
	}
	return internalapidoc.BuildRouteOperationKeyFromHandlerType(handler.Handler.Info.Type)
}

// BuildOperationKeyFromPath returns the path-derived apidoc operation key base
// used for dynamic plugin routes and non-DTO fallback routes.
func BuildOperationKeyFromPath(path string, method string) string {
	return internalapidoc.BuildRouteOperationKeyFromPath(path, method)
}
