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

// dictResolveRequest carries dictionary label resolution parameters.
type dictResolveRequest struct {
	Type         string   `json:"type"`
	Values       []string `json:"values"`
	IncludeLabel bool     `json:"includeLabel"`
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
