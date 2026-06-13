// This file adapts host runtime configuration rows to governed
// plugin-visible host configuration management capability contracts.
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
	capabilityhostconfigcap "lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// Service exposes the runtime-configuration service and management commands.
type runtimeConfigCapabilityService interface {
	capabilityhostconfigcap.AdminService
}

// adapter exposes governed runtime configuration projections.
type runtimeConfigCapabilityAdapter struct {
	tenantFilter tenantspi.PluginTableFilterService
}

var (
	_ capabilityhostconfigcap.AdminService = (*runtimeConfigCapabilityAdapter)(nil)
)

// New creates the host-owned runtime configuration capability adapter.
func newRuntimeConfigCapabilityAdapter(tenantFilter tenantspi.PluginTableFilterService) runtimeConfigCapabilityService {
	return &runtimeConfigCapabilityAdapter{tenantFilter: tenantFilter}
}

// BatchGetRuntimeConfig returns visible runtime configuration projections.
func (a *runtimeConfigCapabilityAdapter) BatchGetRuntimeConfig(ctx context.Context, _ capmodel.CapabilityContext, keys []capabilityhostconfigcap.RuntimeConfigKey) (*capmodel.BatchResult[*capabilityhostconfigcap.RuntimeConfigProjection, capabilityhostconfigcap.RuntimeConfigKey], error) {
	result := &capmodel.BatchResult[*capabilityhostconfigcap.RuntimeConfigProjection, capabilityhostconfigcap.RuntimeConfigKey]{
		Items:      make(map[capabilityhostconfigcap.RuntimeConfigKey]*capabilityhostconfigcap.RuntimeConfigProjection, len(keys)),
		MissingIDs: []capabilityhostconfigcap.RuntimeConfigKey{},
	}
	requestedKeys := make([]string, 0, len(keys))
	requested := make(map[string]capabilityhostconfigcap.RuntimeConfigKey, len(keys))
	for _, key := range keys {
		normalizedKey := strings.TrimSpace(string(key))
		if normalizedKey == "" {
			result.MissingIDs = append(result.MissingIDs, key)
			continue
		}
		if _, exists := requested[normalizedKey]; exists {
			continue
		}
		requested[normalizedKey] = key
		requestedKeys = append(requestedKeys, normalizedKey)
	}
	if len(requestedKeys) == 0 {
		return result, nil
	}

	rows := make([]*entity.SysConfig, 0, len(requestedKeys))
	cols := dao.SysConfig.Columns()
	model := dao.SysConfig.Ctx(ctx).
		Fields(cols.TenantId, cols.Key, cols.Value, cols.Name).
		WhereIn(cols.Key, requestedKeys)
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
	for key, row := range chooseVisibleRuntimeConfigRows(rows, a.currentTenantID(ctx)) {
		requestKey, ok := requested[key]
		if !ok || row == nil {
			continue
		}
		result.Items[requestKey] = &capabilityhostconfigcap.RuntimeConfigProjection{
			Key:       requestKey,
			ValueJSON: []byte(row.Value),
			LabelKey:  "config." + key + ".label",
			Label:     row.Name,
		}
	}
	for _, key := range keys {
		if _, ok := result.Items[key]; !ok && !Contains(result.MissingIDs, key) {
			result.MissingIDs = append(result.MissingIDs, key)
		}
	}
	return result, nil
}

// SetRuntimeConfigJSON writes one visible runtime configuration value and
// advances the runtime-config shared revision after the write succeeds.
func (a *runtimeConfigCapabilityAdapter) SetRuntimeConfigJSON(ctx context.Context, _ capmodel.CapabilityContext, key capabilityhostconfigcap.RuntimeConfigKey, valueJSON []byte) error {
	normalizedKey := strings.TrimSpace(string(key))
	if normalizedKey == "" {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	tenantID := a.currentTenantID(ctx)
	if tenantID < PlatformTenantID {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return dao.SysConfig.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		row, err := a.lockVisibleRow(ctx, tx, normalizedKey, tenantID)
		if err != nil {
			return err
		}
		if row == nil {
			return bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		if _, err = tx.Model(dao.SysConfig.Table()).Safe().Ctx(ctx).
			Where(do.SysConfig{Id: row.Id}).
			Data(do.SysConfig{Value: string(valueJSON)}).
			Update(); err != nil {
			return err
		}
		return BumpSharedRevision(
			ctx,
			tx,
			RuntimeConfigDomain,
			RuntimeConfigGlobalScope,
			RuntimeConfigChangeReason,
		)
	})
}

// currentTenantID returns the active tenant ID for runtime-config queries.
func (a *runtimeConfigCapabilityAdapter) currentTenantID(ctx context.Context) int {
	if a == nil || a.tenantFilter == nil {
		return PlatformTenantID
	}
	return a.tenantFilter.Context(ctx).TenantID
}

// lockVisibleRow locks the tenant-specific row or platform fallback row that
// the current context may update.
func (a *runtimeConfigCapabilityAdapter) lockVisibleRow(ctx context.Context, tx gdb.TX, key string, tenantID int) (*entity.SysConfig, error) {
	var row *entity.SysConfig
	model := tx.Model(dao.SysConfig.Table()).Safe().Ctx(ctx).
		Where(do.SysConfig{Key: key})
	if tenantID > PlatformTenantID {
		model = model.WhereIn(dao.SysConfig.Columns().TenantId, []int{PlatformTenantID, tenantID})
	} else {
		model = model.Where(dao.SysConfig.Columns().TenantId, PlatformTenantID)
	}
	err := model.OrderDesc(dao.SysConfig.Columns().TenantId).LockUpdate().Scan(&row)
	return row, err
}

// chooseVisibleRuntimeConfigRows keeps tenant-specific config rows over platform defaults.
func chooseVisibleRuntimeConfigRows(rows []*entity.SysConfig, tenantID int) map[string]*entity.SysConfig {
	result := make(map[string]*entity.SysConfig, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		existing := result[row.Key]
		if existing == nil || (tenantID > PlatformTenantID && existing.TenantId == PlatformTenantID && row.TenantId == tenantID) {
			result[row.Key] = row
		}
	}
	return result
}
