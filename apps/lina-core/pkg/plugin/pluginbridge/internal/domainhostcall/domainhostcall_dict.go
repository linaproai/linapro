// This file implements the guest-side dict capability hostcall client.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// dictService adapts dictionary reads to host services.
type dictService struct{ baseService }

// Dict creates the dictionary-domain guest client.
func Dict(invoker Invoker) dictcap.Service {
	return dictService{baseService: newBaseService(invoker)}
}

// ResolveLabels resolves visible dictionary labels with opaque missing values.
func (s dictService) ResolveLabels(_ context.Context, _ capmodel.CapabilityContext, input dictcap.ResolveInput) (*capmodel.BatchResult[*dictcap.LabelProjection, dictcap.Value], error) {
	out := &capmodel.BatchResult[*dictcap.LabelProjection, dictcap.Value]{Items: map[dictcap.Value]*dictcap.LabelProjection{}}
	err := s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictResolveLabels, dictResolveRequest{
		Type:         string(input.Type),
		Values:       dictValuesToStrings(input.Values),
		IncludeLabel: input.IncludeLabel,
	}, out)
	return out, err
}

// ListValues returns one bounded page of visible dictionary value candidates.
func (s dictService) ListValues(_ context.Context, _ capmodel.CapabilityContext, input dictcap.ListValuesInput) (*capmodel.PageResult[*dictcap.LabelProjection], error) {
	out := &capmodel.PageResult[*dictcap.LabelProjection]{Items: []*dictcap.LabelProjection{}}
	err := s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictListValues, dictListValuesRequest{
		Type:         string(input.Type),
		Status:       input.Status,
		IncludeLabel: input.IncludeLabel,
		PageNum:      input.Page.PageNum,
		PageSize:     input.Page.PageSize,
	}, out)
	return out, err
}

// EnsureValuesVisible rejects when any requested dictionary value is absent or invisible.
func (s dictService) EnsureValuesVisible(_ context.Context, _ capmodel.CapabilityContext, input dictcap.ResolveInput) error {
	return s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictEnsureValuesVisible, dictResolveRequest{
		Type:         string(input.Type),
		Values:       dictValuesToStrings(input.Values),
		IncludeLabel: input.IncludeLabel,
	}, nil)
}

// dictResolveRequest carries dictionary label resolution parameters.
type dictResolveRequest struct {
	Type         string   `json:"type"`
	Values       []string `json:"values"`
	IncludeLabel bool     `json:"includeLabel"`
}

// dictListValuesRequest carries dictionary candidate listing parameters.
type dictListValuesRequest struct {
	Type         string `json:"type"`
	Status       *int   `json:"status,omitempty"`
	IncludeLabel bool   `json:"includeLabel"`
	PageNum      int    `json:"pageNum,omitempty"`
	PageSize     int    `json:"pageSize,omitempty"`
}

// dictValuesToStrings converts dictionary values to transport strings.
func dictValuesToStrings(values []dictcap.Value) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

var _ dictcap.Service = (*dictService)(nil)
