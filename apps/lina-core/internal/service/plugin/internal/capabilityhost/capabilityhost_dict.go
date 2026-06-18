// This file adapts host dictionary rows to plugin-visible dictionary
// capability contracts and optional runtime labels.
package capabilityhost

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// Service exposes the dictionary domain service and management commands.
type dictCapabilityService interface {
	capabilitydictcap.Service
	capabilitydictcap.AdminService
}

// adapter implements dictionary label projection resolution.
type dictCapabilityAdapter struct {
	tenantFilter tenantspi.PluginTableFilterService
	i18n         i18ncap.Service
}

var (
	_ capabilitydictcap.Service      = (*dictCapabilityAdapter)(nil)
	_ capabilitydictcap.AdminService = (*dictCapabilityAdapter)(nil)
)

// New creates the host-owned dictionary capability adapter.
func newDictCapabilityAdapter(tenantFilter tenantspi.PluginTableFilterService, i18n i18ncap.Service) dictCapabilityService {
	return &dictCapabilityAdapter{tenantFilter: tenantFilter, i18n: i18n}
}

// ResolveLabels resolves visible dictionary labels with opaque missing values.
func (a *dictCapabilityAdapter) ResolveLabels(ctx context.Context, _ capmodel.CapabilityContext, input capabilitydictcap.ResolveInput) (*capmodel.BatchResult[*capabilitydictcap.LabelProjection, capabilitydictcap.Value], error) {
	result := &capmodel.BatchResult[*capabilitydictcap.LabelProjection, capabilitydictcap.Value]{
		Items:      make(map[capabilitydictcap.Value]*capabilitydictcap.LabelProjection, len(input.Values)),
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

	rows := make([]*entity.SysDictData, 0, len(values))
	cols := dao.SysDictData.Columns()
	model := dao.SysDictData.Ctx(ctx).
		Fields(cols.TenantId, cols.DictType, cols.Value, cols.Label, cols.Sort).
		Where(do.SysDictData{DictType: dictType, Status: 1}).
		WhereIn(cols.Value, values).
		OrderAsc(cols.Sort)
	if a != nil && a.tenantFilter != nil {
		tenantID := a.tenantFilter.Context(ctx).TenantID
		if tenantID > PlatformTenantID {
			model = model.WhereIn(cols.TenantId, []int{PlatformTenantID, tenantID})
		} else {
			model = model.Where(cols.TenantId, PlatformTenantID)
		}
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}

	visibleRows := chooseVisibleDictRows(rows, a.currentTenantID(ctx))
	for value, row := range visibleRows {
		requestValue, ok := requested[value]
		if !ok || row == nil {
			continue
		}
		projection := &capabilitydictcap.LabelProjection{
			Type:     input.Type,
			Value:    requestValue,
			LabelKey: dictLabelKey(dictType, value),
		}
		if input.IncludeLabel {
			projection.Label = a.translate(ctx, projection.LabelKey, row.Label)
		}
		result.Items[requestValue] = projection
	}
	for _, value := range input.Values {
		if _, ok := result.Items[value]; !ok && !Contains(result.MissingIDs, value) {
			result.MissingIDs = append(result.MissingIDs, value)
		}
	}
	return result, nil
}

// ListValues returns one bounded page of visible dictionary value candidates.
func (a *dictCapabilityAdapter) ListValues(ctx context.Context, _ capmodel.CapabilityContext, input capabilitydictcap.ListValuesInput) (*capmodel.PageResult[*capabilitydictcap.LabelProjection], error) {
	pageNum, pageSize := NormalizePage(input.Page)
	if pageSize > capabilitydictcap.MaxListValuesPageSize {
		pageSize = capabilitydictcap.MaxListValuesPageSize
	}
	dictType := strings.TrimSpace(string(input.Type))
	if dictType == "" {
		return &capmodel.PageResult[*capabilitydictcap.LabelProjection]{Items: []*capabilitydictcap.LabelProjection{}, Total: 0}, nil
	}

	cols := dao.SysDictData.Columns()
	model := dao.SysDictData.Ctx(ctx).
		Fields(cols.TenantId, cols.DictType, cols.Value, cols.Label, cols.Sort).
		Where(do.SysDictData{DictType: dictType})
	if input.Status != nil {
		model = model.Where(do.SysDictData{Status: *input.Status})
	}
	if a != nil && a.tenantFilter != nil {
		tenantID := a.tenantFilter.Context(ctx).TenantID
		if tenantID > PlatformTenantID {
			model = model.WhereIn(cols.TenantId, []int{PlatformTenantID, tenantID})
		} else {
			model = model.Where(cols.TenantId, PlatformTenantID)
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
		return &capmodel.PageResult[*capabilitydictcap.LabelProjection]{Items: []*capabilitydictcap.LabelProjection{}, Total: total}, nil
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
	visibleRows := chooseVisibleDictRows(rows, a.currentTenantID(ctx))
	items := make([]*capabilitydictcap.LabelProjection, 0, len(pageValues))
	for _, value := range pageValues {
		row := visibleRows[value]
		if row == nil {
			continue
		}
		projection := &capabilitydictcap.LabelProjection{
			Type:     input.Type,
			Value:    capabilitydictcap.Value(row.Value),
			LabelKey: dictLabelKey(dictType, row.Value),
		}
		if input.IncludeLabel {
			projection.Label = a.translate(ctx, projection.LabelKey, row.Label)
		}
		items = append(items, projection)
	}
	return &capmodel.PageResult[*capabilitydictcap.LabelProjection]{Items: items, Total: total}, nil
}

// EnsureValuesVisible rejects when any requested dictionary value is absent or invisible.
func (a *dictCapabilityAdapter) EnsureValuesVisible(ctx context.Context, capCtx capmodel.CapabilityContext, input capabilitydictcap.ResolveInput) error {
	if len(input.Values) > capabilitydictcap.MaxEnsureValuesVisible {
		return bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitydictcap.MaxEnsureValuesVisible))
	}
	result, err := a.ResolveLabels(ctx, capCtx, input)
	if err != nil {
		return err
	}
	if result == nil || len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// Refresh advances the dictionary cache revision for one visible dictionary type.
func (a *dictCapabilityAdapter) Refresh(ctx context.Context, _ capmodel.CapabilityContext, dictType capabilitydictcap.Type) error {
	normalizedType := strings.TrimSpace(string(dictType))
	if normalizedType == "" {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	scope := "type:" + normalizedType
	return dao.SysCacheRevision.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		model := tx.Model(dao.SysDictType.Table()).Safe().Ctx(ctx).
			Where(do.SysDictType{Type: normalizedType})
		if a != nil && a.tenantFilter != nil {
			tenantID := a.tenantFilter.Context(ctx).TenantID
			if tenantID > PlatformTenantID {
				model = model.WhereIn(dao.SysDictType.Columns().TenantId, []int{PlatformTenantID, tenantID})
			} else {
				model = model.Where(dao.SysDictType.Columns().TenantId, PlatformTenantID)
			}
		}
		count, countErr := model.Count()
		if countErr != nil {
			return countErr
		}
		if count == 0 {
			return bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		return BumpSharedRevision(ctx, tx, DictionaryCacheDomain, scope, DictionaryRefreshReason)
	})
}

// chooseVisibleDictRows keeps tenant-specific dictionary rows over platform defaults.
func chooseVisibleDictRows(rows []*entity.SysDictData, tenantID int) map[string]*entity.SysDictData {
	result := make(map[string]*entity.SysDictData, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		existing := result[row.Value]
		if existing == nil || (tenantID > PlatformTenantID && existing.TenantId == PlatformTenantID && row.TenantId == tenantID) {
			result[row.Value] = row
		}
	}
	return result
}

// currentTenantID returns the active tenant ID from the tenant-filter context.
func (a *dictCapabilityAdapter) currentTenantID(ctx context.Context) int {
	if a == nil || a.tenantFilter == nil {
		return PlatformTenantID
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
