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

// Refresh refreshes one governed dictionary type.
func (s dictService) Refresh(_ context.Context, dictType dictcap.Type) error {
	return s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictRefresh, dictTypeRequest{Type: string(dictType)}, nil)
}

// Get returns one visible dictionary type.
func (s dictTypeService) Get(_ context.Context, id int) (*dictcap.TypeInfo, error) {
	var out *dictcap.TypeInfo
	err := s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictTypeGet, dictIDRequest{ID: id}, &out)
	return out, err
}

// BatchGet returns visible dictionary types.
func (s dictTypeService) BatchGet(_ context.Context, ids []int) (*capmodel.BatchResult[*dictcap.TypeInfo, int], error) {
	out := &capmodel.BatchResult[*dictcap.TypeInfo, int]{Items: map[int]*dictcap.TypeInfo{}}
	err := s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictTypeBatchGet, dictIDsRequest{IDs: append([]int(nil), ids...)}, out)
	return out, err
}

// List returns visible dictionary type candidates.
func (s dictTypeService) List(_ context.Context, input dictcap.ListTypesInput) (*capmodel.PageResult[*dictcap.TypeInfo], error) {
	out := &capmodel.PageResult[*dictcap.TypeInfo]{Items: []*dictcap.TypeInfo{}}
	err := s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictTypeList, input, out)
	return out, err
}

// EnsureVisible rejects when any dictionary type ID is absent or invisible.
func (s dictTypeService) EnsureVisible(_ context.Context, ids []int) error {
	return s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictTypeEnsureVisible, dictIDsRequest{IDs: append([]int(nil), ids...)}, nil)
}

// EnsureKeysVisible rejects when any dictionary type key is absent or invisible.
func (s dictTypeService) EnsureKeysVisible(_ context.Context, keys []dictcap.Type) error {
	return s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictTypeEnsureKeysVisible, dictTypeKeysRequest{Keys: dictTypesToStrings(keys)}, nil)
}

// Create creates one dictionary type.
func (s dictTypeService) Create(_ context.Context, input dictcap.CreateTypeInput) (int, error) {
	var out int
	err := s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictTypeCreate, input, &out)
	return out, err
}

// Update updates one visible dictionary type.
func (s dictTypeService) Update(_ context.Context, input dictcap.UpdateTypeInput) error {
	return s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictTypeUpdate, input, nil)
}

// Delete deletes one visible dictionary type.
func (s dictTypeService) Delete(_ context.Context, id int) error {
	return s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictTypeDelete, dictIDRequest{ID: id}, nil)
}

// Get returns one visible dictionary value row.
func (s dictValueService) Get(_ context.Context, id int) (*dictcap.ValueInfo, error) {
	var out *dictcap.ValueInfo
	err := s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictValueGet, dictIDRequest{ID: id}, &out)
	return out, err
}

// BatchGet returns visible dictionary values.
func (s dictValueService) BatchGet(_ context.Context, input dictcap.BatchGetValuesInput) (*capmodel.BatchResult[*dictcap.ValueInfo, dictcap.Value], error) {
	out := &capmodel.BatchResult[*dictcap.ValueInfo, dictcap.Value]{Items: map[dictcap.Value]*dictcap.ValueInfo{}}
	err := s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictValueBatchGet, input, out)
	return out, err
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

// EnsureVisible rejects when any dictionary value row is absent or invisible.
func (s dictValueService) EnsureVisible(_ context.Context, ids []int) error {
	return s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictValueEnsureVisible, dictIDsRequest{IDs: append([]int(nil), ids...)}, nil)
}

// EnsureValuesVisible rejects when any requested dictionary value is absent or invisible.
func (s dictValueService) EnsureValuesVisible(_ context.Context, input dictcap.ResolveInput) error {
	return s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictValueEnsureValuesVisible, dictResolveRequest{
		Type:         string(input.Type),
		Values:       dictValuesToStrings(input.Values),
		IncludeLabel: input.IncludeLabel,
	}, nil)
}

// Create creates one dictionary value.
func (s dictValueService) Create(_ context.Context, input dictcap.CreateValueInput) (int, error) {
	var out int
	err := s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictValueCreate, input, &out)
	return out, err
}

// Update updates one visible dictionary value.
func (s dictValueService) Update(_ context.Context, input dictcap.UpdateValueInput) error {
	return s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictValueUpdate, input, nil)
}

// Delete deletes one visible dictionary value.
func (s dictValueService) Delete(_ context.Context, id int) error {
	return s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictValueDelete, dictIDRequest{ID: id}, nil)
}

// DeleteByType deletes visible dictionary values under one type.
func (s dictValueService) DeleteByType(_ context.Context, dictType dictcap.Type) error {
	return s.callJSONRequest(protocol.HostServiceDict, protocol.HostServiceMethodDictValueDeleteByType, dictTypeRequest{Type: string(dictType)}, nil)
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

type dictIDRequest struct {
	ID int `json:"id"`
}

type dictIDsRequest struct {
	IDs []int `json:"ids"`
}

type dictTypeRequest struct {
	Type string `json:"type"`
}

type dictTypeKeysRequest struct {
	Keys []string `json:"keys"`
}

// dictValuesToStrings converts dictionary values to transport strings.
func dictValuesToStrings(values []dictcap.Value) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

// dictTypesToStrings converts dictionary types to transport strings.
func dictTypesToStrings(values []dictcap.Type) []string {
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
