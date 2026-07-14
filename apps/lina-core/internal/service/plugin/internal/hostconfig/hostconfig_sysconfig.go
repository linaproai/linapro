// This file implements governed sys_config projection, mutation and visibility
// checks for plugin-visible HostConfig services.
package hostconfig

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
	"lina-core/pkg/configvaluetype"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityhostconfigcap "lina-core/pkg/plugin/capability/hostconfigcap"
)

const (
	runtimeConfigCacheDomain  cachecoord.Domain       = "runtime-config"
	runtimeConfigChangeReason cachecoord.ChangeReason = "runtime_config_changed"
)

// Get returns one visible sys_config projection.
func (a *sysConfigCapabilityAdapter) Get(ctx context.Context, key capabilityhostconfigcap.SysConfigKey) (*capabilityhostconfigcap.SysConfigInfo, error) {
	result, err := a.BatchGet(ctx, []capabilityhostconfigcap.SysConfigKey{key})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if projection := result.Items[key]; projection != nil {
		return projection, nil
	}
	return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
}

// BatchGet returns visible sys_config projections.
func (a *sysConfigCapabilityAdapter) BatchGet(ctx context.Context, keys []capabilityhostconfigcap.SysConfigKey) (*capmodel.BatchResult[*capabilityhostconfigcap.SysConfigInfo, capabilityhostconfigcap.SysConfigKey], error) {
	var (
		result = &capmodel.BatchResult[*capabilityhostconfigcap.SysConfigInfo, capabilityhostconfigcap.SysConfigKey]{
			Items:      make(map[capabilityhostconfigcap.SysConfigKey]*capabilityhostconfigcap.SysConfigInfo, len(keys)),
			MissingIDs: []capabilityhostconfigcap.SysConfigKey{},
		}
		requestedKeys = make([]string, 0, len(keys))
		requested     = make(map[string]capabilityhostconfigcap.SysConfigKey, len(keys))
	)
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

	var (
		rows  = make([]*entity.SysConfig, 0, len(requestedKeys))
		cols  = dao.SysConfig.Columns()
		model = dao.SysConfig.Ctx(ctx).
			Fields(cols.TenantId, cols.Key, cols.Value, cols.Name, cols.SystemManageable).
			WhereIn(cols.Key, requestedKeys)
	)
	if a != nil && a.tenantFilter != nil {
		tenantID := a.tenantFilter.Context(ctx).TenantID
		if tenantID > datascope.PlatformTenantID {
			model = model.WhereIn(cols.TenantId, []int{datascope.PlatformTenantID, tenantID})
		} else {
			model = model.Where(cols.TenantId, datascope.PlatformTenantID)
		}
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}
	for key, row := range chooseVisibleSysConfigRows(rows, a.currentTenantID(ctx)) {
		requestKey, ok := requested[key]
		if !ok || row == nil {
			continue
		}
		result.Items[requestKey] = &capabilityhostconfigcap.SysConfigInfo{
			Key:              requestKey,
			Value:            row.Value,
			SystemManageable: row.SystemManageable == 1,
			LabelKey:         "config." + key + ".label",
			Label:            row.Name,
		}
	}
	for _, key := range keys {
		if _, ok := result.Items[key]; !ok && !slices.Contains(result.MissingIDs, key) {
			result.MissingIDs = append(result.MissingIDs, key)
		}
	}
	return result, nil
}

// List returns one bounded page of visible sys_config projections.
func (a *sysConfigCapabilityAdapter) List(ctx context.Context, input capabilityhostconfigcap.ListSysConfigInput) (*capmodel.PageResult[*capabilityhostconfigcap.SysConfigInfo], error) {
	pageNum, pageSize := input.Page.Normalize()
	var (
		cols     = dao.SysConfig.Columns()
		tenantID = a.currentTenantID(ctx)
		model    = dao.SysConfig.Ctx(ctx)
	)
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		model = model.Where(fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", cols.Key, cols.Name), like, like)
	}
	if tenantID > datascope.PlatformTenantID {
		model = model.WhereIn(cols.TenantId, []int{datascope.PlatformTenantID, tenantID})
	} else {
		model = model.Where(cols.TenantId, datascope.PlatformTenantID)
	}
	total, err := model.Clone().Fields(cols.Key).Group(cols.Key).Count()
	if err != nil {
		return nil, err
	}
	keyRows := make([]*struct {
		Key string `orm:"key"`
	}, 0, pageSize)
	if err = model.Clone().
		Fields(cols.Key).
		Group(cols.Key).
		OrderAsc(cols.Key).
		Page(pageNum, pageSize).
		Scan(&keyRows); err != nil {
		return nil, err
	}
	pageKeys := make([]string, 0, len(keyRows))
	for _, row := range keyRows {
		if row != nil && strings.TrimSpace(row.Key) != "" {
			pageKeys = append(pageKeys, row.Key)
		}
	}
	if len(pageKeys) == 0 {
		return &capmodel.PageResult[*capabilityhostconfigcap.SysConfigInfo]{Items: []*capabilityhostconfigcap.SysConfigInfo{}, Total: total}, nil
	}
	rows := make([]*entity.SysConfig, 0, len(pageKeys)*2)
	rowModel := dao.SysConfig.Ctx(ctx).
		Fields(cols.TenantId, cols.Key, cols.Value, cols.Name, cols.SystemManageable).
		WhereIn(cols.Key, pageKeys)
	if tenantID > datascope.PlatformTenantID {
		rowModel = rowModel.WhereIn(cols.TenantId, []int{datascope.PlatformTenantID, tenantID})
	} else {
		rowModel = rowModel.Where(cols.TenantId, datascope.PlatformTenantID)
	}
	if err = rowModel.OrderAsc(cols.Key).OrderDesc(cols.TenantId).Scan(&rows); err != nil {
		return nil, err
	}
	visibleRows := chooseVisibleSysConfigRows(rows, tenantID)
	items := make([]*capabilityhostconfigcap.SysConfigInfo, 0, len(pageKeys))
	for _, key := range pageKeys {
		row := visibleRows[key]
		if row == nil {
			continue
		}
		requestKey := capabilityhostconfigcap.SysConfigKey(key)
		items = append(items, &capabilityhostconfigcap.SysConfigInfo{
			Key:              requestKey,
			Value:            row.Value,
			SystemManageable: row.SystemManageable == 1,
			LabelKey:         "config." + key + ".label",
			Label:            row.Name,
		})
	}
	return &capmodel.PageResult[*capabilityhostconfigcap.SysConfigInfo]{Items: items, Total: total}, nil
}

// SetValue upserts one visible sys_config value. It delegates to BatchSetValue
// so single-key and multi-key writes share transaction and revision semantics.
func (a *sysConfigCapabilityAdapter) SetValue(
	ctx context.Context,
	key capabilityhostconfigcap.SysConfigKey,
	value string,
	options *capabilityhostconfigcap.SetSysConfigValueOptions,
) error {
	return a.BatchSetValue(ctx, []capabilityhostconfigcap.SetSysConfigValueItem{{
		Key:   key,
		Value: value,
	}}, options)
}

// BatchSetValue upserts multiple visible sys_config values in one transaction
// and advances the runtime-config revision once after all writes succeed.
// A missing row is created at the platform tenant scope so plugins can persist
// namespaced settings without install-time host SQL.
//
// options.SystemManageable semantics (shared by every item):
//   - nil options / nil flag on first insert → 0 (plugin closed-loop default)
//   - nil options / nil flag on update → leave existing flag unchanged
//   - non-nil flag → write the flag on insert and update
func (a *sysConfigCapabilityAdapter) BatchSetValue(
	ctx context.Context,
	items []capabilityhostconfigcap.SetSysConfigValueItem,
	options *capabilityhostconfigcap.SetSysConfigValueOptions,
) error {
	if len(items) == 0 {
		return nil
	}
	tenantID := a.currentTenantID(ctx)
	if tenantID < datascope.PlatformTenantID {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if a == nil || a.cacheCoord == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "cachecoord"))
	}
	normalized := make([]capabilityhostconfigcap.SetSysConfigValueItem, 0, len(items))
	for _, item := range items {
		key := strings.TrimSpace(string(item.Key))
		if key == "" {
			return bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		normalized = append(normalized, capabilityhostconfigcap.SetSysConfigValueItem{
			Key:   capabilityhostconfigcap.SysConfigKey(key),
			Value: item.Value,
		})
	}
	systemManageable := (*bool)(nil)
	if options != nil {
		systemManageable = options.SystemManageable
	}
	if err := dao.SysConfig.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		for _, item := range normalized {
			if err := a.upsertSysConfigValue(ctx, string(item.Key), item.Value, tenantID, systemManageable); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return a.markRuntimeConfigChanged(ctx)
}

// upsertSysConfigValue inserts or updates one sys_config row inside an open
// transaction. The name on first insert mirrors the key for admin identity.
func (a *sysConfigCapabilityAdapter) upsertSysConfigValue(
	ctx context.Context,
	key string,
	value string,
	tenantID int,
	systemManageable *bool,
) error {
	row, err := a.lockVisibleRow(ctx, key, tenantID)
	if err != nil {
		return err
	}
	if row == nil {
		_, err = dao.SysConfig.Ctx(ctx).
			Data(do.SysConfig{
				TenantId:         datascope.PlatformTenantID,
				Name:             key,
				Key:              key,
				Value:            value,
				ValueType:        configvaluetype.Text.String(),
				Options:          "",
				IsBuiltin:        0,
				SystemManageable: resolveSystemManageableFlag(systemManageable, 0),
			}).
			Insert()
		return err
	}
	data := do.SysConfig{Value: value}
	if systemManageable != nil {
		data.SystemManageable = resolveSystemManageableFlag(systemManageable, row.SystemManageable)
	}
	_, err = dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{Id: row.Id}).
		Data(data).
		Update()
	return err
}

// resolveSystemManageableFlag maps an optional bool to the SMALLINT storage
// flag. When the pointer is nil, defaultFlag is returned (insert defaults to 0).
func resolveSystemManageableFlag(value *bool, defaultFlag int) int {
	if value == nil {
		return defaultFlag
	}
	if *value {
		return 1
	}
	return 0
}

// Reset clears one visible sys_config value and advances the shared revision.
func (a *sysConfigCapabilityAdapter) Reset(ctx context.Context, key capabilityhostconfigcap.SysConfigKey) error {
	normalizedKey := strings.TrimSpace(string(key))
	if normalizedKey == "" {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	tenantID := a.currentTenantID(ctx)
	if tenantID < datascope.PlatformTenantID {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if a == nil || a.cacheCoord == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "cachecoord"))
	}
	if err := dao.SysConfig.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		row, err := a.lockVisibleRow(ctx, normalizedKey, tenantID)
		if err != nil {
			return err
		}
		if row == nil {
			return bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		if _, err = dao.SysConfig.Ctx(ctx).
			Where(do.SysConfig{Id: row.Id}).
			Data(do.SysConfig{Value: ""}).
			Update(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return a.markRuntimeConfigChanged(ctx)
}

// markRuntimeConfigChanged publishes runtime-config revision changes after
// sys_config writes commit successfully.
func (a *sysConfigCapabilityAdapter) markRuntimeConfigChanged(ctx context.Context) error {
	_, err := a.cacheCoord.MarkChanged(
		ctx,
		runtimeConfigCacheDomain,
		cachecoord.ScopeGlobal,
		runtimeConfigChangeReason,
	)
	return err
}

// EnsureVisible rejects when any sys_config key is absent or invisible.
func (a *sysConfigCapabilityAdapter) EnsureVisible(ctx context.Context, keys []capabilityhostconfigcap.SysConfigKey) error {
	result, err := a.BatchGet(ctx, keys)
	if err != nil {
		return err
	}
	if result == nil || len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// currentTenantID returns the active tenant ID for sys_config queries.
func (a *sysConfigCapabilityAdapter) currentTenantID(ctx context.Context) int {
	if a == nil || a.tenantFilter == nil {
		return datascope.PlatformTenantID
	}
	return a.tenantFilter.Context(ctx).TenantID
}

// lockVisibleRow locks the tenant-specific row or platform fallback row that
// the current context may update.
func (a *sysConfigCapabilityAdapter) lockVisibleRow(ctx context.Context, key string, tenantID int) (*entity.SysConfig, error) {
	cols := dao.SysConfig.Columns()
	var row *entity.SysConfig
	model := dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{Key: key})
	if tenantID > datascope.PlatformTenantID {
		model = model.WhereIn(cols.TenantId, []int{datascope.PlatformTenantID, tenantID})
	} else {
		model = model.Where(cols.TenantId, datascope.PlatformTenantID)
	}
	err := model.OrderDesc(cols.TenantId).LockUpdate().Scan(&row)
	return row, err
}

// chooseVisibleSysConfigRows keeps tenant-specific config rows over platform defaults.
func chooseVisibleSysConfigRows(rows []*entity.SysConfig, tenantID int) map[string]*entity.SysConfig {
	result := make(map[string]*entity.SysConfig, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		existing := result[row.Key]
		if existing == nil || (tenantID > datascope.PlatformTenantID && existing.TenantId == datascope.PlatformTenantID && row.TenantId == tenantID) {
			result[row.Key] = row
		}
	}
	return result
}

// Get reports that no sys_config backend was injected.
func (sysConfigUnavailableService) Get(context.Context, capabilityhostconfigcap.SysConfigKey) (*capabilityhostconfigcap.SysConfigInfo, error) {
	return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "sys_config"))
}

// BatchGet reports that no sys_config backend was injected.
func (sysConfigUnavailableService) BatchGet(context.Context, []capabilityhostconfigcap.SysConfigKey) (*capmodel.BatchResult[*capabilityhostconfigcap.SysConfigInfo, capabilityhostconfigcap.SysConfigKey], error) {
	return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "sys_config"))
}

// List reports that no sys_config backend was injected.
func (sysConfigUnavailableService) List(context.Context, capabilityhostconfigcap.ListSysConfigInput) (*capmodel.PageResult[*capabilityhostconfigcap.SysConfigInfo], error) {
	return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "sys_config"))
}

// SetValue reports that no sys_config backend was injected.
func (sysConfigUnavailableService) SetValue(context.Context, capabilityhostconfigcap.SysConfigKey, string, *capabilityhostconfigcap.SetSysConfigValueOptions) error {
	return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "sys_config"))
}

// BatchSetValue reports that no sys_config backend was injected.
func (sysConfigUnavailableService) BatchSetValue(context.Context, []capabilityhostconfigcap.SetSysConfigValueItem, *capabilityhostconfigcap.SetSysConfigValueOptions) error {
	return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "sys_config"))
}

// Reset reports that no sys_config backend was injected.
func (sysConfigUnavailableService) Reset(context.Context, capabilityhostconfigcap.SysConfigKey) error {
	return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "sys_config"))
}

// EnsureVisible reports that no sys_config backend was injected.
func (sysConfigUnavailableService) EnsureVisible(context.Context, []capabilityhostconfigcap.SysConfigKey) error {
	return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "sys_config"))
}
