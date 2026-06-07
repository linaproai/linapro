// Package configcap adapts host runtime configuration rows to governed
// plugin-visible runtime configuration capability contracts.
package configcap

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
	capabilityconfigcap "lina-core/pkg/plugin/capability/configcap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// Service exposes the runtime-configuration service and management commands.
type Service interface {
	capabilityconfigcap.Service
	capabilityconfigcap.AdminService
}

// adapter exposes governed runtime configuration projections.
type adapter struct {
	tenantFilter tenantcap.PluginTableFilterService
}

var (
	_ capabilityconfigcap.Service      = (*adapter)(nil)
	_ capabilityconfigcap.AdminService = (*adapter)(nil)
)

// New creates the host-owned runtime configuration capability adapter.
func New(tenantFilter tenantcap.PluginTableFilterService) Service {
	return &adapter{tenantFilter: tenantFilter}
}

// BatchGetConfig returns visible runtime configuration projections.
func (a *adapter) BatchGetConfig(ctx context.Context, _ capmodel.CapabilityContext, keys []capabilityconfigcap.ConfigKey) (*capmodel.BatchResult[*capabilityconfigcap.Projection, capabilityconfigcap.ConfigKey], error) {
	result := &capmodel.BatchResult[*capabilityconfigcap.Projection, capabilityconfigcap.ConfigKey]{
		Items:      make(map[capabilityconfigcap.ConfigKey]*capabilityconfigcap.Projection, len(keys)),
		MissingIDs: []capabilityconfigcap.ConfigKey{},
	}
	requestedKeys := make([]string, 0, len(keys))
	requested := make(map[string]capabilityconfigcap.ConfigKey, len(keys))
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
		if tenantID > domaincap.PlatformTenantID {
			model = model.WhereIn(cols.TenantId, []int{domaincap.PlatformTenantID, tenantID})
		} else {
			model = model.Where(cols.TenantId, domaincap.PlatformTenantID)
		}
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}
	for key, row := range chooseVisibleRows(rows, a.currentTenantID(ctx)) {
		requestKey, ok := requested[key]
		if !ok || row == nil {
			continue
		}
		result.Items[requestKey] = &capabilityconfigcap.Projection{
			Key:       requestKey,
			ValueJSON: []byte(row.Value),
			LabelKey:  "config." + key + ".label",
			Label:     row.Name,
		}
	}
	for _, key := range keys {
		if _, ok := result.Items[key]; !ok && !domaincap.Contains(result.MissingIDs, key) {
			result.MissingIDs = append(result.MissingIDs, key)
		}
	}
	return result, nil
}

// SetConfigJSON writes one visible runtime configuration value and advances the
// runtime-config shared revision after the write succeeds.
func (a *adapter) SetConfigJSON(ctx context.Context, _ capmodel.CapabilityContext, key capabilityconfigcap.ConfigKey, valueJSON []byte) error {
	normalizedKey := strings.TrimSpace(string(key))
	if normalizedKey == "" {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	tenantID := a.currentTenantID(ctx)
	if tenantID < domaincap.PlatformTenantID {
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
		return domaincap.BumpSharedRevision(
			ctx,
			tx,
			domaincap.RuntimeConfigDomain,
			domaincap.RuntimeConfigGlobalScope,
			domaincap.RuntimeConfigChangeReason,
		)
	})
}

// currentTenantID returns the active tenant ID for runtime-config queries.
func (a *adapter) currentTenantID(ctx context.Context) int {
	if a == nil || a.tenantFilter == nil {
		return domaincap.PlatformTenantID
	}
	return a.tenantFilter.Context(ctx).TenantID
}

// lockVisibleRow locks the tenant-specific row or platform fallback row that
// the current context may update.
func (a *adapter) lockVisibleRow(ctx context.Context, tx gdb.TX, key string, tenantID int) (*entity.SysConfig, error) {
	var row *entity.SysConfig
	model := tx.Model(dao.SysConfig.Table()).Safe().Ctx(ctx).
		Where(do.SysConfig{Key: key})
	if tenantID > domaincap.PlatformTenantID {
		model = model.WhereIn(dao.SysConfig.Columns().TenantId, []int{domaincap.PlatformTenantID, tenantID})
	} else {
		model = model.Where(dao.SysConfig.Columns().TenantId, domaincap.PlatformTenantID)
	}
	err := model.OrderDesc(dao.SysConfig.Columns().TenantId).LockUpdate().Scan(&row)
	return row, err
}

// chooseVisibleRows keeps tenant-specific config rows over platform defaults.
func chooseVisibleRows(rows []*entity.SysConfig, tenantID int) map[string]*entity.SysConfig {
	result := make(map[string]*entity.SysConfig, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		existing := result[row.Key]
		if existing == nil || (tenantID > domaincap.PlatformTenantID && existing.TenantId == domaincap.PlatformTenantID && row.TenantId == tenantID) {
			result[row.Key] = row
		}
	}
	return result
}
