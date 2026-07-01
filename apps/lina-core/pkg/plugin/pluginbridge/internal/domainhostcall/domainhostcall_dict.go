// This file implements the guest-side dict capability hostcall client.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/statusflag"
)

// dictService adapts dictionary subresource reads to host services.
type dictService struct{ baseService }

// dictTypeService reports unavailable dynamic type-management methods.
type dictTypeService struct{ baseService }

// dictValueService adapts dictionary value reads to host services.
type dictValueService struct{ baseService }

// Dict creates the dictionary-domain guest client.
func Dict(invoker Invoker) dictcap.Service {
	return dictService{baseService: newBaseService(invoker)}
}

// Type returns dictionary type subresource methods.
func (s dictService) Type() dictcap.TypeService {
	return dictTypeService{baseService: s.baseService}
}

// Value returns dictionary value subresource methods.
func (s dictService) Value() dictcap.ValueService {
	return dictValueService{baseService: s.baseService}
}

// Refresh is not published as a dynamic dict host-service method.
func (s dictService) Refresh(context.Context, dictcap.Type) error {
	return unsupportedDynamicMethodError("dict.refresh")
}

// Get is not published as a dynamic dictionary type host-service method.
func (s dictTypeService) Get(context.Context, int) (*dictcap.TypeInfo, error) {
	return nil, unsupportedDynamicMethodError("dict.type.get")
}

// BatchGet is not published as a dynamic dictionary type host-service method.
func (s dictTypeService) BatchGet(context.Context, []int) (*capmodel.BatchResult[*dictcap.TypeInfo, int], error) {
	return nil, unsupportedDynamicMethodError("dict.type.batch_get")
}

// List is not published as a dynamic dictionary type host-service method.
func (s dictTypeService) List(context.Context, dictcap.ListTypesInput) (*capmodel.PageResult[*dictcap.TypeInfo], error) {
	return nil, unsupportedDynamicMethodError("dict.type.list")
}

// EnsureVisible is not published as a dynamic dictionary type host-service method.
func (s dictTypeService) EnsureVisible(context.Context, []int) error {
	return unsupportedDynamicMethodError("dict.type.visible.ensure")
}

// EnsureKeysVisible is not published as a dynamic dictionary type host-service method.
func (s dictTypeService) EnsureKeysVisible(context.Context, []dictcap.Type) error {
	return unsupportedDynamicMethodError("dict.type.keys.visible.ensure")
}

// Create is not published as a dynamic dictionary type host-service method.
func (s dictTypeService) Create(context.Context, dictcap.CreateTypeInput) (int, error) {
	return 0, unsupportedDynamicMethodError("dict.type.create")
}

// Update is not published as a dynamic dictionary type host-service method.
func (s dictTypeService) Update(context.Context, dictcap.UpdateTypeInput) error {
	return unsupportedDynamicMethodError("dict.type.update")
}

// Delete is not published as a dynamic dictionary type host-service method.
func (s dictTypeService) Delete(context.Context, int) error {
	return unsupportedDynamicMethodError("dict.type.delete")
}

// Get is not published as a dynamic dictionary value host-service method.
func (s dictValueService) Get(context.Context, int) (*dictcap.ValueInfo, error) {
	return nil, unsupportedDynamicMethodError("dict.value.get")
}

// BatchGet delegates visible value resolution to the registered label resolver.
func (s dictValueService) BatchGet(ctx context.Context, input dictcap.BatchGetValuesInput) (*capmodel.BatchResult[*dictcap.ValueInfo, dictcap.Value], error) {
	labels, err := s.ResolveLabels(ctx, dictcap.ResolveInput{
		Type:         input.Type,
		Values:       input.Values,
		IncludeLabel: input.IncludeLabel,
	})
	if err != nil || labels == nil {
		return nil, err
	}
	out := &capmodel.BatchResult[*dictcap.ValueInfo, dictcap.Value]{
		Items:      map[dictcap.Value]*dictcap.ValueInfo{},
		MissingIDs: append([]dictcap.Value(nil), labels.MissingIDs...),
	}
	for value, label := range labels.Items {
		if label == nil {
			continue
		}
		out.Items[value] = &dictcap.ValueInfo{
			Type:     label.Type,
			Value:    label.Value,
			LabelKey: label.LabelKey,
			Label:    label.Label,
		}
	}
	return out, nil
}

// ResolveLabels resolves visible dictionary labels with opaque missing values.
func (s dictValueService) ResolveLabels(_ context.Context, input dictcap.ResolveInput) (*capmodel.BatchResult[*dictcap.LabelInfo, dictcap.Value], error) {
	out := &capmodel.BatchResult[*dictcap.LabelInfo, dictcap.Value]{Items: map[dictcap.Value]*dictcap.LabelInfo{}}
	err := s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictValueResolveLabels, dictResolveRequest{
		Type:         string(input.Type),
		Values:       dictValuesToStrings(input.Values),
		IncludeLabel: input.IncludeLabel,
	}, out)
	return out, err
}

// List returns one bounded page of visible dictionary value candidates.
func (s dictValueService) List(_ context.Context, input dictcap.ListValuesInput) (*capmodel.PageResult[*dictcap.ValueInfo], error) {
	out := &capmodel.PageResult[*dictcap.ValueInfo]{Items: []*dictcap.ValueInfo{}}
	err := s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictListValues, dictListValuesRequest{
		Type:         string(input.Type),
		Status:       dictStatusIntPtr(input.Status),
		IncludeLabel: input.IncludeLabel,
		PageNum:      input.Page.PageNum,
		PageSize:     input.Page.PageSize,
	}, out)
	return out, err
}

// dictStatusIntPtr converts a shared status flag into the wire integer pointer.
func dictStatusIntPtr(status *statusflag.Enabled) *int {
	if status == nil {
		return nil
	}
	value := int(*status)
	return &value
}

// EnsureVisible is not published as a dynamic dictionary value row host-service method.
func (s dictValueService) EnsureVisible(context.Context, []int) error {
	return unsupportedDynamicMethodError("dict.value.visible.ensure")
}

// EnsureValuesVisible rejects when any requested dictionary value is absent or invisible.
func (s dictValueService) EnsureValuesVisible(_ context.Context, input dictcap.ResolveInput) error {
	return s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictValueEnsureValuesVisible, dictResolveRequest{
		Type:         string(input.Type),
		Values:       dictValuesToStrings(input.Values),
		IncludeLabel: input.IncludeLabel,
	}, nil)
}

// Create is not published as a dynamic dictionary value host-service method.
func (s dictValueService) Create(context.Context, dictcap.CreateValueInput) (int, error) {
	return 0, unsupportedDynamicMethodError("dict.value.create")
}

// Update is not published as a dynamic dictionary value host-service method.
func (s dictValueService) Update(context.Context, dictcap.UpdateValueInput) error {
	return unsupportedDynamicMethodError("dict.value.update")
}

// Delete is not published as a dynamic dictionary value host-service method.
func (s dictValueService) Delete(context.Context, int) error {
	return unsupportedDynamicMethodError("dict.value.delete")
}

// DeleteByType is not published as a dynamic dictionary value host-service method.
func (s dictValueService) DeleteByType(context.Context, dictcap.Type) error {
	return unsupportedDynamicMethodError("dict.value.delete_by_type")
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

var (
	_ dictcap.Service      = (*dictService)(nil)
	_ dictcap.TypeService  = (*dictTypeService)(nil)
	_ dictcap.ValueService = (*dictValueService)(nil)
)
