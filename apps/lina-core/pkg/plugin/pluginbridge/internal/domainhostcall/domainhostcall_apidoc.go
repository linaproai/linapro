// This file implements guest-side API-documentation capability hostcall
// clients. It keeps route text DTO conversion next to the apidoc domain logic.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// apiDocService adapts API documentation capability calls to host services.
type apiDocService struct{ baseService }

// APIDoc creates the API-documentation localization guest client.
func APIDoc(invoker Invoker) apidoccap.Service {
	return apiDocService{baseService: newBaseService(invoker)}
}

// ResolveRouteText resolves one route's localized module tag and operation summary.
func (s apiDocService) ResolveRouteText(_ context.Context, input apidoccap.RouteTextInput) apidoccap.RouteTextOutput {
	var out apidoccap.RouteTextOutput
	if err := s.callJSONRequest(protocol.HostServiceAPIDoc, protocol.HostServiceMethodAPIDocResolveRouteText, apiDocRouteTextRequestFromInput(input), &out); err != nil {
		return apidoccap.RouteTextOutput{}
	}
	return out
}

// ResolveRouteTexts resolves multiple route texts with one apidoc catalog load.
func (s apiDocService) ResolveRouteTexts(_ context.Context, inputs []apidoccap.RouteTextInput) []apidoccap.RouteTextOutput {
	request := apiDocRouteTextsRequest{Inputs: make([]apiDocRouteTextRequest, 0, len(inputs))}
	for _, input := range inputs {
		request.Inputs = append(request.Inputs, apiDocRouteTextRequestFromInput(input))
	}
	var out []apidoccap.RouteTextOutput
	if err := s.callJSONRequest(protocol.HostServiceAPIDoc, protocol.HostServiceMethodAPIDocResolveRouteTexts, request, &out); err != nil {
		return nil
	}
	return out
}

// FindRouteTitleOperationKeys finds localized module tag operation keys by keyword.
func (s apiDocService) FindRouteTitleOperationKeys(_ context.Context, keyword string) []string {
	var out []string
	if err := s.callJSONRequest(protocol.HostServiceAPIDoc, protocol.HostServiceMethodAPIDocFindRouteTitleOperationKeys, keywordRequest{Keyword: keyword}, &out); err != nil {
		return nil
	}
	return out
}

// keywordRequest carries one keyword for JSON capability methods.
type keywordRequest struct {
	Keyword string `json:"keyword"`
}

// apiDocRouteTextRequest carries one API route text lookup.
type apiDocRouteTextRequest struct {
	OperationKey    string `json:"operationKey"`
	Method          string `json:"method"`
	Path            string `json:"path"`
	FallbackTitle   string `json:"fallbackTitle"`
	FallbackSummary string `json:"fallbackSummary"`
}

// apiDocRouteTextsRequest carries a batch of API route text lookups.
type apiDocRouteTextsRequest struct {
	Inputs []apiDocRouteTextRequest `json:"inputs"`
}

// apiDocRouteTextRequestFromInput converts the public capability input to JSON DTO.
func apiDocRouteTextRequestFromInput(input apidoccap.RouteTextInput) apiDocRouteTextRequest {
	return apiDocRouteTextRequest{
		OperationKey:    input.OperationKey,
		Method:          input.Method,
		Path:            input.Path,
		FallbackTitle:   input.FallbackTitle,
		FallbackSummary: input.FallbackSummary,
	}
}

var _ apidoccap.Service = (*apiDocService)(nil)
