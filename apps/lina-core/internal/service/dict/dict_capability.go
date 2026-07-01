// This file adapts host dictionary rows to plugin-visible dictionary
// capability contracts and optional runtime labels.
package dict

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/statusflag"
)

const (
	dictionaryCacheDomain   cachecoord.Domain       = "dictionary"
	dictionaryRefreshReason cachecoord.ChangeReason = "dictionary_refreshed"
)

// adapter implements dictionary label projection resolution.
type dictCapabilityAdapter struct {
	tenantFilter tenantcap.FilterService
	i18n         i18ncap.Service
	cacheCoord   cachecoord.Service
}

var (
	_ capabilitydictcap.Service      = (*dictCapabilityAdapter)(nil)
	_ capabilitydictcap.TypeService  = (*dictTypeCapabilityAdapter)(nil)
	_ capabilitydictcap.ValueService = (*dictValueCapabilityAdapter)(nil)
)

// NewCapabilityAdapter creates the host-owned dictionary capability adapter.
func NewCapabilityAdapter(
	tenantFilter tenantcap.FilterService,
	i18n i18ncap.Service,
	cacheCoord cachecoord.Service,
) capabilitydictcap.Service {
	return &dictCapabilityAdapter{tenantFilter: tenantFilter, i18n: i18n, cacheCoord: cacheCoord}
}

// Type returns governed dictionary type subresource methods.
func (a *dictCapabilityAdapter) Type() capabilitydictcap.TypeService {
	return &dictTypeCapabilityAdapter{parent: a}
}

// Value returns governed dictionary value subresource methods.
func (a *dictCapabilityAdapter) Value() capabilitydictcap.ValueService {
	return &dictValueCapabilityAdapter{parent: a}
}

// dictTypeCapabilityAdapter implements dictionary type methods.
type dictTypeCapabilityAdapter struct {
	parent *dictCapabilityAdapter
}

// dictValueCapabilityAdapter implements dictionary value methods.
type dictValueCapabilityAdapter struct {
	parent *dictCapabilityAdapter
}

// Get returns one visible dictionary type.
func (a *dictTypeCapabilityAdapter) Get(ctx context.Context, id int) (*capabilitydictcap.TypeInfo, error) {
	result, err := a.BatchGet(ctx, []int{id})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if projection := result.Items[id]; projection != nil {
		return projection, nil
	}
	return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
}

// BatchGet returns visible dictionary types and opaque missing IDs.
func (a *dictTypeCapabilityAdapter) BatchGet(ctx context.Context, ids []int) (*capmodel.BatchResult[*capabilitydictcap.TypeInfo, int], error) {
	result := &capmodel.BatchResult[*capabilitydictcap.TypeInfo, int]{
		Items:      make(map[int]*capabilitydictcap.TypeInfo, len(ids)),
		MissingIDs: []int{},
	}
	if len(ids) > capabilitydictcap.MaxBatchGetTypes {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitydictcap.MaxBatchGetTypes))
	}
	requested := make(map[int]struct{}, len(ids))
	requestIDs := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			result.MissingIDs = append(result.MissingIDs, id)
			continue
		}
		if _, ok := requested[id]; ok {
			continue
		}
		requested[id] = struct{}{}
		requestIDs = append(requestIDs, id)
	}
	if len(requestIDs) == 0 {
		return result, nil
	}
	var (
		rows  = make([]*entity.SysDictType, 0, len(requestIDs))
		cols  = dao.SysDictType.Columns()
		model = dao.SysDictType.Ctx(ctx).
			Fields(cols.Id, cols.TenantId, cols.Type, cols.Name, cols.Status).
			WhereIn(cols.Id, requestIDs)
	)
	if a != nil && a.parent != nil && a.parent.tenantFilter != nil {
		model = applyDictTenantFilter(ctx, model, a.parent.tenantFilter, cols.TenantId)
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		result.Items[row.Id] = a.projectType(ctx, row, false)
	}
	for _, id := range ids {
		if _, ok := result.Items[id]; !ok && !slices.Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// List returns bounded visible dictionary type candidates.
func (a *dictTypeCapabilityAdapter) List(ctx context.Context, input capabilitydictcap.ListTypesInput) (*capmodel.PageResult[*capabilitydictcap.TypeInfo], error) {
	pageNum, pageSize := input.Page.Normalize()
	if pageSize > capabilitydictcap.MaxListTypesPageSize {
		pageSize = capabilitydictcap.MaxListTypesPageSize
	}
	cols := dao.SysDictType.Columns()
	model := dao.SysDictType.Ctx(ctx).
		Fields(cols.Id, cols.TenantId, cols.Type, cols.Name, cols.Status)
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		model = model.Where(fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", cols.Type, cols.Name), like, like)
	}
	if dictType := strings.TrimSpace(string(input.Type)); dictType != "" {
		model = model.Where(cols.Type, dictType)
	}
	if input.Status != nil {
		model = model.Where(cols.Status, *input.Status)
	}
	if a != nil && a.parent != nil && a.parent.tenantFilter != nil {
		model = applyDictTenantFilter(ctx, model, a.parent.tenantFilter, cols.TenantId)
	}
	total, err := model.Clone().Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.SysDictType, 0, pageSize)
	if err = model.Clone().Page(pageNum, pageSize).OrderDesc(cols.Id).Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*capabilitydictcap.TypeInfo, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, a.projectType(ctx, row, input.IncludeLabel))
	}
	return &capmodel.PageResult[*capabilitydictcap.TypeInfo]{Items: items, Total: total}, nil
}

// EnsureVisible rejects when any requested type ID is absent or invisible.
func (a *dictTypeCapabilityAdapter) EnsureVisible(ctx context.Context, ids []int) error {
	result, err := a.BatchGet(ctx, ids)
	if err != nil {
		return err
	}
	if result == nil || len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// EnsureKeysVisible rejects when any requested type key is absent or invisible.
func (a *dictTypeCapabilityAdapter) EnsureKeysVisible(ctx context.Context, keys []capabilitydictcap.Type) error {
	if len(keys) > capabilitydictcap.MaxEnsureTypeKeysVisible {
		return bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitydictcap.MaxEnsureTypeKeysVisible))
	}
	normalized := make([]string, 0, len(keys))
	for _, key := range keys {
		value := strings.TrimSpace(string(key))
		if value == "" {
			return bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		normalized = append(normalized, value)
	}
	cols := dao.SysDictType.Columns()
	model := dao.SysDictType.Ctx(ctx).Fields(cols.Type).WhereIn(cols.Type, normalized)
	if a != nil && a.parent != nil && a.parent.tenantFilter != nil {
		model = applyDictTenantFilter(ctx, model, a.parent.tenantFilter, cols.TenantId)
	}
	rows := make([]*entity.SysDictType, 0, len(normalized))
	if err := model.Scan(&rows); err != nil {
		return err
	}
	visible := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row != nil {
			visible[row.Type] = struct{}{}
		}
	}
	for _, key := range normalized {
		if _, ok := visible[key]; !ok {
			return bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
	}
	return nil
}

// Create creates one dictionary type.
func (a *dictTypeCapabilityAdapter) Create(ctx context.Context, input capabilitydictcap.CreateTypeInput) (int, error) {
	id, err := dao.SysDictType.Ctx(ctx).Data(do.SysDictType{
		Type:   strings.TrimSpace(string(input.Type)),
		Name:   input.Name,
		Status: input.Status,
		Remark: input.Remark,
	}).InsertAndGetId()
	return int(id), err
}

// Update updates one visible dictionary type.
func (a *dictTypeCapabilityAdapter) Update(ctx context.Context, input capabilitydictcap.UpdateTypeInput) error {
	if err := a.EnsureVisible(ctx, []int{input.ID}); err != nil {
		return err
	}
	data := do.SysDictType{}
	if input.Type != nil {
		data.Type = strings.TrimSpace(string(*input.Type))
	}
	if input.Name != nil {
		data.Name = *input.Name
	}
	if input.Status != nil {
		data.Status = *input.Status
	}
	if input.Remark != nil {
		data.Remark = *input.Remark
	}
	_, err := dao.SysDictType.Ctx(ctx).Where(do.SysDictType{Id: input.ID}).Data(data).Update()
	return err
}

// Delete deletes one visible dictionary type.
func (a *dictTypeCapabilityAdapter) Delete(ctx context.Context, id int) error {
	if err := a.EnsureVisible(ctx, []int{id}); err != nil {
		return err
	}
	_, err := dao.SysDictType.Ctx(ctx).Where(do.SysDictType{Id: id}).Delete()
	return err
}

// projectType converts a host dictionary type row.
func (a *dictTypeCapabilityAdapter) projectType(ctx context.Context, row *entity.SysDictType, includeLabel bool) *capabilitydictcap.TypeInfo {
	if row == nil {
		return nil
	}
	projection := &capabilitydictcap.TypeInfo{
		ID:       row.Id,
		Type:     capabilitydictcap.Type(row.Type),
		Name:     row.Name,
		Status:   statusflag.Enabled(row.Status),
		LabelKey: "dict." + row.Type + ".label",
	}
	if includeLabel && a != nil && a.parent != nil {
		projection.Label = a.parent.translate(ctx, projection.LabelKey, row.Name)
	}
	return projection
}

// Get returns one visible dictionary value by row ID.
func (a *dictValueCapabilityAdapter) Get(ctx context.Context, id int) (*capabilitydictcap.ValueInfo, error) {
	if err := a.EnsureVisible(ctx, []int{id}); err != nil {
		return nil, err
	}
	var row *entity.SysDictData
	cols := dao.SysDictData.Columns()
	model := dao.SysDictData.Ctx(ctx).
		Fields(cols.Id, cols.TenantId, cols.DictType, cols.Value, cols.Label, cols.Sort, cols.Status).
		Where(do.SysDictData{Id: id})
	if a != nil && a.parent != nil && a.parent.tenantFilter != nil {
		model = applyDictTenantFilter(ctx, model, a.parent.tenantFilter, cols.TenantId)
	}
	if err := model.Scan(&row); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return a.projectValue(ctx, row, true), nil
}

// BatchGet returns visible dictionary values by type and value.
func (a *dictValueCapabilityAdapter) BatchGet(ctx context.Context, input capabilitydictcap.BatchGetValuesInput) (*capmodel.BatchResult[*capabilitydictcap.ValueInfo, capabilitydictcap.Value], error) {
	if len(input.Values) > capabilitydictcap.MaxBatchGetValues {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitydictcap.MaxBatchGetValues))
	}
	result, err := a.ResolveLabels(ctx, capabilitydictcap.ResolveInput{
		Type:         input.Type,
		Values:       input.Values,
		IncludeLabel: input.IncludeLabel,
	})
	if err != nil {
		return nil, err
	}
	out := &capmodel.BatchResult[*capabilitydictcap.ValueInfo, capabilitydictcap.Value]{
		Items:      make(map[capabilitydictcap.Value]*capabilitydictcap.ValueInfo, len(result.Items)),
		MissingIDs: append([]capabilitydictcap.Value(nil), result.MissingIDs...),
	}
	for value, projection := range result.Items {
		if projection == nil {
			continue
		}
		out.Items[value] = &capabilitydictcap.ValueInfo{
			Type:     projection.Type,
			Value:    projection.Value,
			LabelKey: projection.LabelKey,
			Label:    projection.Label,
		}
	}
	return out, nil
}

// ResolveLabels resolves visible dictionary labels with opaque missing values.
func (a *dictValueCapabilityAdapter) ResolveLabels(ctx context.Context, input capabilitydictcap.ResolveInput) (*capmodel.BatchResult[*capabilitydictcap.LabelInfo, capabilitydictcap.Value], error) {
	result := &capmodel.BatchResult[*capabilitydictcap.LabelInfo, capabilitydictcap.Value]{
		Items:      make(map[capabilitydictcap.Value]*capabilitydictcap.LabelInfo, len(input.Values)),
		MissingIDs: []capabilitydictcap.Value{},
	}
	dictType := strings.TrimSpace(string(input.Type))
	if dictType == "" || len(input.Values) == 0 {
		for _, value := range input.Values {
			result.MissingIDs = append(result.MissingIDs, value)
		}
		return result, nil
	}

	values := make([]string, 0, len(input.Values))
	requested := make(map[string]capabilitydictcap.Value, len(input.Values))
	for _, value := range input.Values {
		normalizedValue := strings.TrimSpace(string(value))
		if normalizedValue == "" {
			result.MissingIDs = append(result.MissingIDs, value)
			continue
		}
		if _, exists := requested[normalizedValue]; exists {
			continue
		}
		requested[normalizedValue] = value
		values = append(values, normalizedValue)
	}
	if len(values) == 0 {
		return result, nil
	}

	var (
		rows  = make([]*entity.SysDictData, 0, len(values))
		cols  = dao.SysDictData.Columns()
		model = dao.SysDictData.Ctx(ctx).
			Fields(cols.TenantId, cols.DictType, cols.Value, cols.Label, cols.Sort).
			Where(do.SysDictData{DictType: dictType, Status: 1}).
			WhereIn(cols.Value, values).
			OrderAsc(cols.Sort)
	)
	if a != nil && a.parent != nil && a.parent.tenantFilter != nil {
		tenantID := a.parent.tenantFilter.Context(ctx).TenantID
		if tenantID > datascope.PlatformTenantID {
			model = model.WhereIn(cols.TenantId, []int{datascope.PlatformTenantID, tenantID})
		} else {
			model = model.Where(cols.TenantId, datascope.PlatformTenantID)
		}
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}

	visibleRows := chooseVisibleDictRows(rows, a.parent.currentTenantID(ctx))
	for value, row := range visibleRows {
		requestValue, ok := requested[value]
		if !ok || row == nil {
			continue
		}
		projection := &capabilitydictcap.LabelInfo{
			Type:     input.Type,
			Value:    requestValue,
			LabelKey: dictLabelKey(dictType, value),
		}
		if input.IncludeLabel {
			projection.Label = a.parent.translate(ctx, projection.LabelKey, row.Label)
		}
		result.Items[requestValue] = projection
	}
	for _, value := range input.Values {
		if _, ok := result.Items[value]; !ok && !slices.Contains(result.MissingIDs, value) {
			result.MissingIDs = append(result.MissingIDs, value)
		}
	}
	return result, nil
}

// List returns one bounded page of visible dictionary values.
func (a *dictValueCapabilityAdapter) List(ctx context.Context, input capabilitydictcap.ListValuesInput) (*capmodel.PageResult[*capabilitydictcap.ValueInfo], error) {
	pageNum, pageSize := input.Page.Normalize()
	if pageSize > capabilitydictcap.MaxListValuesPageSize {
		pageSize = capabilitydictcap.MaxListValuesPageSize
	}
	dictType := strings.TrimSpace(string(input.Type))
	if dictType == "" {
		return &capmodel.PageResult[*capabilitydictcap.ValueInfo]{Items: []*capabilitydictcap.ValueInfo{}, Total: 0}, nil
	}

	cols := dao.SysDictData.Columns()
	model := dao.SysDictData.Ctx(ctx).
		Fields(cols.TenantId, cols.DictType, cols.Value, cols.Label, cols.Sort).
		Where(do.SysDictData{DictType: dictType})
	if input.Status != nil {
		model = model.Where(do.SysDictData{Status: *input.Status})
	}
	if a != nil && a.parent != nil && a.parent.tenantFilter != nil {
		tenantID := a.parent.tenantFilter.Context(ctx).TenantID
		if tenantID > datascope.PlatformTenantID {
			model = model.WhereIn(cols.TenantId, []int{datascope.PlatformTenantID, tenantID})
		} else {
			model = model.Where(cols.TenantId, datascope.PlatformTenantID)
		}
	}
	total, err := model.Clone().Fields(cols.Value).Group(cols.Value).Count()
	if err != nil {
		return nil, err
	}
	valueRows := make([]*struct {
		Value string `orm:"value"`
	}, 0, pageSize)
	if err = model.Clone().
		Fields(cols.Value, "MIN("+cols.Sort+") AS min_sort").
		Group(cols.Value).
		OrderAsc("min_sort").
		OrderAsc(cols.Value).
		Page(pageNum, pageSize).
		Scan(&valueRows); err != nil {
		return nil, err
	}
	pageValues := make([]string, 0, len(valueRows))
	for _, row := range valueRows {
		if row != nil && strings.TrimSpace(row.Value) != "" {
			pageValues = append(pageValues, row.Value)
		}
	}
	if len(pageValues) == 0 {
		return &capmodel.PageResult[*capabilitydictcap.ValueInfo]{Items: []*capabilitydictcap.ValueInfo{}, Total: total}, nil
	}

	rows := make([]*entity.SysDictData, 0, len(pageValues)*2)
	if err = model.Clone().
		Fields(cols.TenantId, cols.DictType, cols.Value, cols.Label, cols.Sort).
		WhereIn(cols.Value, pageValues).
		OrderAsc(cols.Sort).
		OrderAsc(cols.Value).
		Scan(&rows); err != nil {
		return nil, err
	}
	visibleRows := chooseVisibleDictRows(rows, a.parent.currentTenantID(ctx))
	items := make([]*capabilitydictcap.ValueInfo, 0, len(pageValues))
	for _, value := range pageValues {
		row := visibleRows[value]
		if row == nil {
			continue
		}
		items = append(items, a.projectValue(ctx, row, input.IncludeLabel))
	}
	return &capmodel.PageResult[*capabilitydictcap.ValueInfo]{Items: items, Total: total}, nil
}

// EnsureVisible rejects when any requested dictionary value row is absent or invisible.
func (a *dictValueCapabilityAdapter) EnsureVisible(ctx context.Context, ids []int) error {
	requested := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		requested = append(requested, id)
	}
	if len(requested) == 0 {
		return nil
	}
	cols := dao.SysDictData.Columns()
	model := dao.SysDictData.Ctx(ctx).Fields(cols.Id).WhereIn(cols.Id, requested)
	if a != nil && a.parent != nil && a.parent.tenantFilter != nil {
		model = applyDictTenantFilter(ctx, model, a.parent.tenantFilter, cols.TenantId)
	}
	rows := make([]*entity.SysDictData, 0, len(requested))
	if err := model.Scan(&rows); err != nil {
		return err
	}
	if len(rows) != len(requested) {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// EnsureValuesVisible rejects when any requested dictionary value is absent or invisible.
func (a *dictValueCapabilityAdapter) EnsureValuesVisible(ctx context.Context, input capabilitydictcap.ResolveInput) error {
	if len(input.Values) > capabilitydictcap.MaxEnsureValuesVisible {
		return bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitydictcap.MaxEnsureValuesVisible))
	}
	result, err := a.ResolveLabels(ctx, input)
	if err != nil {
		return err
	}
	if result == nil || len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// Create creates one dictionary value.
func (a *dictValueCapabilityAdapter) Create(ctx context.Context, input capabilitydictcap.CreateValueInput) (int, error) {
	id, err := dao.SysDictData.Ctx(ctx).Data(do.SysDictData{
		DictType: strings.TrimSpace(string(input.Type)),
		Value:    strings.TrimSpace(string(input.Value)),
		Label:    input.Label,
		Sort:     input.Sort,
		TagStyle: input.TagStyle,
		CssClass: input.CssClass,
		Status:   input.Status,
		Remark:   input.Remark,
	}).InsertAndGetId()
	return int(id), err
}

// Update updates one visible dictionary value.
func (a *dictValueCapabilityAdapter) Update(ctx context.Context, input capabilitydictcap.UpdateValueInput) error {
	if err := a.EnsureVisible(ctx, []int{input.ID}); err != nil {
		return err
	}
	data := do.SysDictData{}
	if input.Type != nil {
		data.DictType = strings.TrimSpace(string(*input.Type))
	}
	if input.Value != nil {
		data.Value = strings.TrimSpace(string(*input.Value))
	}
	if input.Label != nil {
		data.Label = *input.Label
	}
	if input.Sort != nil {
		data.Sort = *input.Sort
	}
	if input.TagStyle != nil {
		data.TagStyle = *input.TagStyle
	}
	if input.CssClass != nil {
		data.CssClass = *input.CssClass
	}
	if input.Status != nil {
		data.Status = *input.Status
	}
	if input.Remark != nil {
		data.Remark = *input.Remark
	}
	_, err := dao.SysDictData.Ctx(ctx).Where(do.SysDictData{Id: input.ID}).Data(data).Update()
	return err
}

// Delete deletes one visible dictionary value.
func (a *dictValueCapabilityAdapter) Delete(ctx context.Context, id int) error {
	if err := a.EnsureVisible(ctx, []int{id}); err != nil {
		return err
	}
	_, err := dao.SysDictData.Ctx(ctx).Where(do.SysDictData{Id: id}).Delete()
	return err
}

// DeleteByType deletes values under one visible dictionary type.
func (a *dictValueCapabilityAdapter) DeleteByType(ctx context.Context, dictType capabilitydictcap.Type) error {
	typeSvc := &dictTypeCapabilityAdapter{parent: a.parent}
	if err := typeSvc.EnsureKeysVisible(ctx, []capabilitydictcap.Type{dictType}); err != nil {
		return err
	}
	_, err := dao.SysDictData.Ctx(ctx).Where(do.SysDictData{DictType: strings.TrimSpace(string(dictType))}).Delete()
	return err
}

// projectValue converts a host dictionary data row.
func (a *dictValueCapabilityAdapter) projectValue(ctx context.Context, row *entity.SysDictData, includeLabel bool) *capabilitydictcap.ValueInfo {
	if row == nil {
		return nil
	}
	projection := &capabilitydictcap.ValueInfo{
		ID:       row.Id,
		Type:     capabilitydictcap.Type(row.DictType),
		Value:    capabilitydictcap.Value(row.Value),
		LabelKey: dictLabelKey(row.DictType, row.Value),
		Sort:     row.Sort,
		Status:   statusflag.Enabled(row.Status),
	}
	if includeLabel && a != nil && a.parent != nil {
		projection.Label = a.parent.translate(ctx, projection.LabelKey, row.Label)
	}
	return projection
}

// Refresh advances the dictionary cache revision for one visible dictionary type.
func (a *dictCapabilityAdapter) Refresh(ctx context.Context, dictType capabilitydictcap.Type) error {
	normalizedType := strings.TrimSpace(string(dictType))
	if normalizedType == "" {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if a == nil || a.cacheCoord == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "cachecoord"))
	}
	var (
		scope = "type:" + normalizedType
		cols  = dao.SysDictType.Columns()
		model = dao.SysDictType.Ctx(ctx).
			Where(do.SysDictType{Type: normalizedType})
	)
	if a.tenantFilter != nil {
		tenantID := a.tenantFilter.Context(ctx).TenantID
		if tenantID > datascope.PlatformTenantID {
			model = model.WhereIn(cols.TenantId, []int{datascope.PlatformTenantID, tenantID})
		} else {
			model = model.Where(cols.TenantId, datascope.PlatformTenantID)
		}
	}
	count, err := model.Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	_, err = a.cacheCoord.MarkChanged(ctx, dictionaryCacheDomain, cachecoord.Scope(scope), dictionaryRefreshReason)
	return err
}

// chooseVisibleDictRows keeps tenant-specific dictionary rows over platform defaults.
func chooseVisibleDictRows(rows []*entity.SysDictData, tenantID int) map[string]*entity.SysDictData {
	result := make(map[string]*entity.SysDictData, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		existing := result[row.Value]
		if existing == nil || (tenantID > datascope.PlatformTenantID && existing.TenantId == datascope.PlatformTenantID && row.TenantId == tenantID) {
			result[row.Value] = row
		}
	}
	return result
}

// currentTenantID returns the active tenant ID from the tenant-filter context.
func (a *dictCapabilityAdapter) currentTenantID(ctx context.Context) int {
	if a == nil || a.tenantFilter == nil {
		return datascope.PlatformTenantID
	}
	return a.tenantFilter.Context(ctx).TenantID
}

// translate resolves one dictionary label when the caller requested a label.
func (a *dictCapabilityAdapter) translate(ctx context.Context, key string, fallback string) string {
	if a == nil || a.i18n == nil {
		return fallback
	}
	return a.i18n.Translate(ctx, key, fallback)
}

// dictLabelKey builds the stable runtime dictionary label key.
func dictLabelKey(dictType string, value string) string {
	return "dict." + dictType + "." + value + ".label"
}

// applyDictTenantFilter applies platform fallback and tenant override visibility.
func applyDictTenantFilter(ctx context.Context, model *gdb.Model, tenantFilter tenantcap.FilterService, tenantColumn string) *gdb.Model {
	tenantID := tenantFilter.Context(ctx).TenantID
	if tenantID > datascope.PlatformTenantID {
		return model.WhereIn(tenantColumn, []int{datascope.PlatformTenantID, tenantID})
	}
	return model.Where(tenantColumn, datascope.PlatformTenantID)
}
