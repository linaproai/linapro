// Package dictcap adapts host dictionary rows to plugin-visible dictionary
// capability contracts and optional runtime labels.
package dictcap

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/hostservices/internal/domaincap"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// Service exposes the dictionary domain service and management commands.
type Service interface {
	capabilitydictcap.Service
	capabilitydictcap.AdminService
}

// adapter implements dictionary label projection resolution.
type adapter struct {
	tenantFilter tenantcap.PluginTableFilterService
	i18n         i18ncap.Service
}

var (
	_ capabilitydictcap.Service      = (*adapter)(nil)
	_ capabilitydictcap.AdminService = (*adapter)(nil)
)

// New creates the host-owned dictionary capability adapter.
func New(tenantFilter tenantcap.PluginTableFilterService, i18n i18ncap.Service) Service {
	return &adapter{tenantFilter: tenantFilter, i18n: i18n}
}

// ResolveLabels resolves visible dictionary labels with opaque missing values.
func (a *adapter) ResolveLabels(ctx context.Context, _ capmodel.CapabilityContext, input capabilitydictcap.ResolveInput) (*capmodel.BatchResult[*capabilitydictcap.LabelProjection, capabilitydictcap.Value], error) {
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
		if tenantID > domaincap.PlatformTenantID {
			model = model.WhereIn(cols.TenantId, []int{domaincap.PlatformTenantID, tenantID})
		} else {
			model = model.Where(cols.TenantId, domaincap.PlatformTenantID)
		}
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}

	visibleRows := chooseVisibleRows(rows, a.currentTenantID(ctx))
	for value, row := range visibleRows {
		requestValue, ok := requested[value]
		if !ok || row == nil {
			continue
		}
		projection := &capabilitydictcap.LabelProjection{
			Type:     input.Type,
			Value:    requestValue,
			LabelKey: labelKey(dictType, value),
		}
		if input.IncludeLabel {
			projection.Label = a.translate(ctx, projection.LabelKey, row.Label)
		}
		result.Items[requestValue] = projection
	}
	for _, value := range input.Values {
		if _, ok := result.Items[value]; !ok && !domaincap.Contains(result.MissingIDs, value) {
			result.MissingIDs = append(result.MissingIDs, value)
		}
	}
	return result, nil
}

// Refresh advances the dictionary cache revision for one visible dictionary type.
func (a *adapter) Refresh(ctx context.Context, _ capmodel.CapabilityContext, dictType capabilitydictcap.Type) error {
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
			if tenantID > domaincap.PlatformTenantID {
				model = model.WhereIn(dao.SysDictType.Columns().TenantId, []int{domaincap.PlatformTenantID, tenantID})
			} else {
				model = model.Where(dao.SysDictType.Columns().TenantId, domaincap.PlatformTenantID)
			}
		}
		count, countErr := model.Count()
		if countErr != nil {
			return countErr
		}
		if count == 0 {
			return bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		return domaincap.BumpSharedRevision(ctx, tx, domaincap.DictionaryCacheDomain, scope, domaincap.DictionaryRefreshReason)
	})
}

// chooseVisibleRows keeps tenant-specific dictionary rows over platform defaults.
func chooseVisibleRows(rows []*entity.SysDictData, tenantID int) map[string]*entity.SysDictData {
	result := make(map[string]*entity.SysDictData, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		existing := result[row.Value]
		if existing == nil || (tenantID > domaincap.PlatformTenantID && existing.TenantId == domaincap.PlatformTenantID && row.TenantId == tenantID) {
			result[row.Value] = row
		}
	}
	return result
}

// currentTenantID returns the active tenant ID from the tenant-filter context.
func (a *adapter) currentTenantID(ctx context.Context) int {
	if a == nil || a.tenantFilter == nil {
		return domaincap.PlatformTenantID
	}
	return a.tenantFilter.Context(ctx).TenantID
}

// translate resolves one dictionary label when the caller requested a label.
func (a *adapter) translate(ctx context.Context, key string, fallback string) string {
	if a == nil || a.i18n == nil {
		return fallback
	}
	return a.i18n.Translate(ctx, key, fallback)
}

// labelKey builds the stable runtime dictionary label key.
func labelKey(dictType string, value string) string {
	return "dict." + dictType + "." + value + ".label"
}
