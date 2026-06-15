// This file adapts host plugin registry and tenant enablement storage
// to plugin-visible plugin governance capability contracts.
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
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
)

const (
	pluginInstalledYes            = 1
	pluginStatusEnabled           = 1
	pluginScopeNatureTenantAware  = "tenant_aware"
	pluginInstallModeTenantScoped = "tenant_scoped"
	tenantEnablementStateKey      = "__tenant_enabled__"
	tenantPluginEnabledValue      = "enabled"
	tenantPluginDisabledValue     = "disabled"
)

// Service exposes plugin governance services, scoped views, and management commands.
type pluginCapabilityService interface {
	capabilityplugincap.Service
	capabilityplugincap.AdminService
	// ForPlugin returns a plugin-bound plugin-domain namespace.
	ForPlugin(pluginID string) capabilityplugincap.Service
}

// adapter implements the plugin-domain namespace and tenant enablement commands.
type pluginCapabilityAdapter struct {
	pluginID      string
	configFactory capabilityplugincap.ConfigServiceFactory
	state         capabilityplugincap.StateService
	lifecycle     capabilityplugincap.LifecycleService
}

var (
	_ capabilityplugincap.Service      = (*pluginCapabilityAdapter)(nil)
	_ capabilityplugincap.AdminService = (*pluginCapabilityAdapter)(nil)
)

// New creates the host-owned plugin governance capability adapter.
func newPluginCapabilityAdapter(
	configFactory capabilityplugincap.ConfigServiceFactory,
	state capabilityplugincap.StateService,
	lifecycle capabilityplugincap.LifecycleService,
) pluginCapabilityService {
	return &pluginCapabilityAdapter{configFactory: configFactory, state: state, lifecycle: lifecycle}
}

// ForPlugin returns a plugin-bound plugin-domain namespace.
func (a *pluginCapabilityAdapter) ForPlugin(pluginID string) capabilityplugincap.Service {
	if a == nil {
		return nil
	}
	clone := *a
	clone.pluginID = strings.TrimSpace(pluginID)
	return &clone
}

// Config returns the current plugin's static configuration reader.
func (a *pluginCapabilityAdapter) Config() capabilityplugincap.ConfigService {
	if a == nil || a.configFactory == nil || strings.TrimSpace(a.pluginID) == "" {
		return nil
	}
	return a.configFactory.ForPlugin(a.pluginID)
}

// State returns plugin state and provider enablement lookups.
func (a *pluginCapabilityAdapter) State() capabilityplugincap.StateService {
	if a == nil {
		return nil
	}
	return a.state
}

// Lifecycle returns plugin lifecycle orchestration operations.
func (a *pluginCapabilityAdapter) Lifecycle() capabilityplugincap.LifecycleService {
	if a == nil {
		return nil
	}
	return a.lifecycle
}

// Registry returns the plugin governance projection service.
func (a *pluginCapabilityAdapter) Registry() capabilityplugincap.RegistryService {
	if a == nil {
		return nil
	}
	return a
}

// BatchGet returns visible plugin projections and opaque missing IDs.
func (a *pluginCapabilityAdapter) BatchGet(ctx context.Context, _ capmodel.CapabilityContext, ids []capabilityplugincap.PluginID) (*capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID], error) {
	result := &capmodel.BatchResult[*capabilityplugincap.Projection, capabilityplugincap.PluginID]{
		Items:      make(map[capabilityplugincap.PluginID]*capabilityplugincap.Projection, len(ids)),
		MissingIDs: []capabilityplugincap.PluginID{},
	}
	if len(ids) == 0 {
		return result, nil
	}
	requestedIDs := make([]string, 0, len(ids))
	requested := make(map[string]capabilityplugincap.PluginID, len(ids))
	for _, id := range ids {
		normalizedID := strings.TrimSpace(string(id))
		if normalizedID == "" {
			result.MissingIDs = append(result.MissingIDs, id)
			continue
		}
		if _, ok := requested[normalizedID]; ok {
			continue
		}
		requested[normalizedID] = id
		requestedIDs = append(requestedIDs, normalizedID)
	}
	if len(requestedIDs) == 0 {
		return result, nil
	}

	rows := make([]*entity.SysPlugin, 0, len(requestedIDs))
	cols := dao.SysPlugin.Columns()
	if err := dao.SysPlugin.Ctx(ctx).
		Fields(cols.PluginId, cols.Version, cols.Installed, cols.Status, cols.CurrentState).
		WhereIn(cols.PluginId, requestedIDs).
		Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		requestID, ok := requested[row.PluginId]
		if !ok {
			continue
		}
		result.Items[requestID] = &capabilityplugincap.Projection{
			ID:        capabilityplugincap.PluginID(row.PluginId),
			Version:   row.Version,
			Installed: row.Installed == pluginInstalledYes,
			Enabled:   row.Status == pluginStatusEnabled,
			Status:    row.CurrentState,
		}
	}
	for _, id := range ids {
		if _, ok := result.Items[id]; !ok && !Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// ListTenantPlugins returns tenant-controllable plugins with current tenant enablement.
func (a *pluginCapabilityAdapter) ListTenantPlugins(ctx context.Context, capCtx capmodel.CapabilityContext) (*capmodel.PageResult[*capabilityplugincap.TenantProjection], error) {
	tenantID, err := TenantID(capCtx.TenantID)
	if err != nil || tenantID <= PlatformTenantID {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	rows := make([]*entity.SysPlugin, 0)
	if err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{
			Installed:   pluginInstalledYes,
			Status:      pluginStatusEnabled,
			ScopeNature: pluginScopeNatureTenantAware,
			InstallMode: pluginInstallModeTenantScoped,
		}).
		OrderAsc(dao.SysPlugin.Columns().PluginId).
		Scan(&rows); err != nil {
		return nil, err
	}
	states, err := a.tenantEnabledStates(ctx, tenantID, pluginIDsFromRows(rows))
	if err != nil {
		return nil, err
	}
	items := make([]*capabilityplugincap.TenantProjection, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, tenantPluginProjection(row, states[row.PluginId]))
	}
	return &capmodel.PageResult[*capabilityplugincap.TenantProjection]{Items: items, Total: len(items)}, nil
}

// SetEnabled changes tenant plugin enablement after target checks.
func (a *pluginCapabilityAdapter) SetEnabled(ctx context.Context, capCtx capmodel.CapabilityContext, id capabilityplugincap.PluginID, enabled bool) error {
	tenantID, err := TenantID(capCtx.TenantID)
	if err != nil || tenantID <= PlatformTenantID {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	pluginID := strings.TrimSpace(string(id))
	if pluginID == "" {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if err = a.ensureTenantScopedPlugin(ctx, pluginID); err != nil {
		return err
	}
	return a.setTenantPluginEnabled(ctx, tenantID, pluginID, enabled, false)
}

// ProvisionTenantDefaults creates missing default tenant plugin enablement rows.
func (a *pluginCapabilityAdapter) ProvisionTenantDefaults(ctx context.Context, _ capmodel.CapabilityContext, tenantIDValue capmodel.DomainID) error {
	tenantID, err := TenantID(tenantIDValue)
	if err != nil || tenantID <= PlatformTenantID {
		return nil
	}
	rows := make([]*entity.SysPlugin, 0)
	if err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{
			Installed:               pluginInstalledYes,
			Status:                  pluginStatusEnabled,
			ScopeNature:             pluginScopeNatureTenantAware,
			InstallMode:             pluginInstallModeTenantScoped,
			AutoEnableForNewTenants: true,
		}).
		Scan(&rows); err != nil {
		return err
	}
	for _, row := range rows {
		if row == nil || strings.TrimSpace(row.PluginId) == "" {
			continue
		}
		if err = a.setTenantPluginEnabled(ctx, tenantID, row.PluginId, true, true); err != nil {
			return err
		}
	}
	return nil
}

// tenantEnabledStates returns the tenant state map for the requested plugins.
func (a *pluginCapabilityAdapter) tenantEnabledStates(ctx context.Context, tenantID int, pluginIDs []string) (map[string]bool, error) {
	result := make(map[string]bool, len(pluginIDs))
	if len(pluginIDs) == 0 {
		return result, nil
	}
	rows := make([]*entity.SysPluginState, 0, len(pluginIDs))
	if err := dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{TenantId: tenantID, StateKey: tenantEnablementStateKey}).
		WhereIn(dao.SysPluginState.Columns().PluginId, pluginIDs).
		Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row != nil {
			result[row.PluginId] = row.Enabled
		}
	}
	return result, nil
}

// ensureTenantScopedPlugin verifies the plugin can be controlled per tenant.
func (a *pluginCapabilityAdapter) ensureTenantScopedPlugin(ctx context.Context, pluginID string) error {
	count, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{
			PluginId:    pluginID,
			Installed:   pluginInstalledYes,
			Status:      pluginStatusEnabled,
			ScopeNature: pluginScopeNatureTenantAware,
			InstallMode: pluginInstallModeTenantScoped,
		}).
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// setTenantPluginEnabled upserts one tenant plugin state and bumps runtime cache revision.
func (a *pluginCapabilityAdapter) setTenantPluginEnabled(ctx context.Context, tenantID int, pluginID string, enabled bool, insertOnly bool) error {
	identity := do.SysPluginState{
		PluginId: pluginID,
		TenantId: tenantID,
		StateKey: tenantEnablementStateKey,
	}
	return dao.SysPluginState.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, insertErr := tx.Model(dao.SysPluginState.Table()).Safe().Ctx(ctx).Data(do.SysPluginState{
			PluginId:   identity.PluginId,
			TenantId:   identity.TenantId,
			StateKey:   identity.StateKey,
			StateValue: pluginEnablementStateValue(enabled),
			Enabled:    enabled,
		}).InsertIgnore()
		if insertErr != nil {
			return insertErr
		}
		affected, affectedErr := result.RowsAffected()
		if affectedErr != nil {
			return affectedErr
		}
		if insertOnly {
			if affected == 0 {
				return nil
			}
			return bumpPluginRuntimeCacheRevision(ctx, tx)
		}
		_, updateErr := tx.Model(dao.SysPluginState.Table()).Safe().Ctx(ctx).
			Where(identity).
			Data(do.SysPluginState{
				StateValue: pluginEnablementStateValue(enabled),
				Enabled:    enabled,
			}).
			Update()
		if updateErr != nil {
			return updateErr
		}
		return bumpPluginRuntimeCacheRevision(ctx, tx)
	})
}

// bumpPluginRuntimeCacheRevision advances the shared plugin-runtime revision.
func bumpPluginRuntimeCacheRevision(ctx context.Context, tx gdb.TX) error {
	return BumpSharedRevision(
		ctx,
		tx,
		PluginRuntimeCacheDomain,
		PluginRuntimeCacheScopeGlobal,
		TenantPluginRuntimeChangeReason,
	)
}

// pluginIDsFromRows collects non-empty plugin IDs from registry rows.
func pluginIDsFromRows(rows []*entity.SysPlugin) []string {
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		if row == nil || strings.TrimSpace(row.PluginId) == "" {
			continue
		}
		result = append(result, row.PluginId)
	}
	return result
}

// tenantPluginProjection converts a host plugin registry row into a stable projection.
func tenantPluginProjection(row *entity.SysPlugin, tenantEnabled bool) *capabilityplugincap.TenantProjection {
	return &capabilityplugincap.TenantProjection{
		ID:            capabilityplugincap.PluginID(row.PluginId),
		Name:          row.Name,
		Version:       row.Version,
		Type:          row.Type,
		Description:   row.Remark,
		Installed:     row.Installed == pluginInstalledYes,
		Enabled:       row.Status == pluginStatusEnabled,
		ScopeNature:   row.ScopeNature,
		InstallMode:   row.InstallMode,
		TenantEnabled: tenantEnabled,
	}
}

// pluginEnablementStateValue converts the tenant enablement flag to persisted text.
func pluginEnablementStateValue(enabled bool) string {
	if enabled {
		return tenantPluginEnabledValue
	}
	return tenantPluginDisabledValue
}
