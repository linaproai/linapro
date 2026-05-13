// Package tenantplugin implements tenant-scoped plugin enablement governance.
package tenantplugin

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	plugincontract "lina-core/pkg/pluginservice/contract"
	"lina-plugin-multi-tenant/backend/internal/dao"
	"lina-plugin-multi-tenant/backend/internal/model/do"
	"lina-plugin-multi-tenant/backend/internal/model/entity"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

const (
	// pluginInstalledYes is the host sys_plugin installed flag for installed plugins.
	pluginInstalledYes = 1
	// pluginStatusEnabled is the host sys_plugin status flag for enabled plugins.
	pluginStatusEnabled = 1
	// pluginTypeSource is the host sys_plugin type for source plugins.
	pluginTypeSource = "source"
	// pluginHostStateEnabled is the stable enabled lifecycle state name.
	pluginHostStateEnabled = "enabled"
	// pluginScopeNaturePlatformOnly marks platform-only plugins.
	pluginScopeNaturePlatformOnly = "platform_only"
	// pluginScopeNatureTenantAware marks tenant-aware plugins.
	pluginScopeNatureTenantAware = "tenant_aware"
	// pluginInstallModeGlobal marks globally enabled plugins.
	pluginInstallModeGlobal = "global"
	// pluginInstallModeTenantScoped marks tenant-controlled plugins.
	pluginInstallModeTenantScoped = "tenant_scoped"
	// tenantEnablementStateKey is the sys_plugin_state key for tenant plugin enablement.
	tenantEnablementStateKey = "__tenant_enabled__"
	// tenantPluginEnabledValue stores enabled tenant plugin state for diagnostics.
	tenantPluginEnabledValue = "enabled"
	// tenantPluginDisabledValue stores disabled tenant plugin state for diagnostics.
	tenantPluginDisabledValue = "disabled"
	// tableSysCacheRevision is the host shared cache-revision table.
	tableSysCacheRevision = "sys_cache_revision"
	// pluginRuntimeCacheDomain coordinates plugin runtime and menu derived caches.
	pluginRuntimeCacheDomain = "plugin-runtime"
	// pluginRuntimeCacheScopeGlobal invalidates every plugin-runtime cache scope.
	pluginRuntimeCacheScopeGlobal = "global"
	// tenantPluginRuntimeChangeReason records tenant plugin enablement changes.
	tenantPluginRuntimeChangeReason = "tenant_plugin_enablement_changed"
)

// Service defines tenant plugin-governance operations.
type Service interface {
	// List returns tenant-controllable plugins with current tenant enablement.
	List(ctx context.Context) (*ListOutput, error)
	// SetEnabled updates one tenant plugin enablement row.
	SetEnabled(ctx context.Context, pluginID string, enabled bool) error
	// ProvisionForTenant provisions default tenant plugin enablement.
	ProvisionForTenant(ctx context.Context, tenantID int64) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc plugincontract.BizCtxService
}

// New creates and returns a tenant plugin governance service.
func New(bizCtxSvc plugincontract.BizCtxService) Service {
	return &serviceImpl{bizCtxSvc: bizCtxSvc}
}

// Entity is the tenant plugin-governance projection.
type Entity struct {
	Id            string
	Name          string
	Version       string
	Type          string
	Description   string
	Installed     int
	Enabled       int
	ScopeNature   string
	InstallMode   string
	TenantEnabled int
}

// ListOutput defines tenant plugin list output.
type ListOutput struct {
	List  []*Entity
	Total int
}

// pluginRuntimeCacheRevisionDO is the local DO payload for sys_cache_revision writes.
type pluginRuntimeCacheRevisionDO struct {
	g.Meta   `orm:"table:sys_cache_revision, do:true"`
	Id       any // Primary key ID
	TenantId any // Revision tenant scope, 0 means platform/global
	Domain   any // Cache domain
	Scope    any // Cache invalidation scope
	Revision any // Monotonic shared revision
	Reason   any // Latest change reason
}

// pluginRuntimeCacheRevisionRow projects one sys_cache_revision row.
type pluginRuntimeCacheRevisionRow struct {
	Id       int64 `json:"id" orm:"id"`
	Revision int64 `json:"revision" orm:"revision"`
}

// List returns tenant-controllable plugins with current tenant enablement.
func (s *serviceImpl) List(ctx context.Context) (*ListOutput, error) {
	tenantID, err := s.requireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []*entity.SysPlugin
	err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{
			Installed:   pluginInstalledYes,
			Status:      pluginStatusEnabled,
			ScopeNature: pluginScopeNatureTenantAware,
			InstallMode: pluginInstallModeTenantScoped,
		}).
		OrderAsc(dao.SysPlugin.Columns().PluginId).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	list := make([]*Entity, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		enabled, err := s.tenantEnabled(ctx, row.PluginId, tenantID)
		if err != nil {
			return nil, err
		}
		list = append(list, pluginEntity(row, enabled))
	}
	return &ListOutput{List: list, Total: len(list)}, nil
}

// SetEnabled updates one tenant plugin enablement row.
func (s *serviceImpl) SetEnabled(ctx context.Context, pluginID string, enabled bool) error {
	tenantID, err := s.requireTenantID(ctx)
	if err != nil {
		return err
	}
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return bizerr.NewCode(CodePluginNotFound)
	}
	if err = s.ensureTenantScopedPlugin(ctx, normalizedPluginID); err != nil {
		return err
	}
	return s.setTenantPluginEnabled(ctx, tenantID, normalizedPluginID, enabled)
}

// ProvisionForTenant provisions default tenant-scoped plugin enablement for a
// newly created tenant. This is explicit domain orchestration rather than a
// lifecycle event subscription: the project does not keep an outbox table or an
// unfinished cross-plugin event bus for tenant creation.
func (s *serviceImpl) ProvisionForTenant(ctx context.Context, tenantID int64) error {
	if tenantID <= shared.PlatformTenantID {
		return nil
	}
	var rows []*entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{
			Installed:               pluginInstalledYes,
			Status:                  pluginStatusEnabled,
			ScopeNature:             pluginScopeNatureTenantAware,
			InstallMode:             pluginInstallModeTenantScoped,
			AutoEnableForNewTenants: true,
		}).
		Scan(&rows)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row == nil || strings.TrimSpace(row.PluginId) == "" {
			continue
		}
		if err = s.setTenantPluginEnabled(ctx, tenantID, row.PluginId, true); err != nil {
			return err
		}
	}
	return nil
}

// setTenantPluginEnabled upserts one tenant plugin enablement row.
func (s *serviceImpl) setTenantPluginEnabled(ctx context.Context, tenantID int64, pluginID string, enabled bool) error {
	identity := do.SysPluginState{
		PluginId: pluginID,
		TenantId: tenantID,
		StateKey: tenantEnablementStateKey,
	}
	return dao.SysPluginState.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, insertErr := tx.Model(dao.SysPluginState.Table()).Safe().Ctx(ctx).Data(do.SysPluginState{
			PluginId:   identity.PluginId,
			TenantId:   identity.TenantId,
			StateKey:   identity.StateKey,
			StateValue: enablementStateValue(enabled),
			Enabled:    enabled,
		}).InsertIgnore()
		if insertErr != nil {
			return insertErr
		}
		_, updateErr := tx.Model(dao.SysPluginState.Table()).Safe().Ctx(ctx).
			Where(identity).
			Data(do.SysPluginState{
				StateValue: enablementStateValue(enabled),
				Enabled:    enabled,
			}).
			Update()
		if updateErr != nil {
			return updateErr
		}
		_, revisionErr := s.bumpPluginRuntimeCacheRevision(ctx, tx)
		return revisionErr
	})
}

// bumpPluginRuntimeCacheRevision publishes the host plugin-runtime cache revision.
func (s *serviceImpl) bumpPluginRuntimeCacheRevision(ctx context.Context, tx gdb.TX) (int64, error) {
	_, err := tx.Model(tableSysCacheRevision).Safe().Ctx(ctx).Data(pluginRuntimeCacheRevisionDO{
		TenantId: shared.PlatformTenantID,
		Domain:   pluginRuntimeCacheDomain,
		Scope:    pluginRuntimeCacheScopeGlobal,
		Revision: 0,
		Reason:   tenantPluginRuntimeChangeReason,
	}).InsertIgnore()
	if err != nil {
		return 0, err
	}

	var row *pluginRuntimeCacheRevisionRow
	err = tx.Model(tableSysCacheRevision).Safe().Ctx(ctx).
		Fields("id", "revision").
		Where(pluginRuntimeCacheRevisionDO{
			TenantId: shared.PlatformTenantID,
			Domain:   pluginRuntimeCacheDomain,
			Scope:    pluginRuntimeCacheScopeGlobal,
		}).
		LockUpdate().
		Scan(&row)
	if err != nil {
		return 0, err
	}
	if row == nil {
		return 0, bizerr.NewCode(CodePluginRuntimeRevisionUnavailable)
	}

	revision := row.Revision + 1
	_, err = tx.Model(tableSysCacheRevision).Safe().Ctx(ctx).
		Where(pluginRuntimeCacheRevisionDO{Id: row.Id}).
		Data(pluginRuntimeCacheRevisionDO{
			Revision: revision,
			Reason:   tenantPluginRuntimeChangeReason,
		}).
		Update()
	if err != nil {
		return 0, err
	}
	return revision, nil
}

// requireTenantID returns the current request tenant id or a stable bizerr.
func (s *serviceImpl) requireTenantID(ctx context.Context) (int64, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	tenantID := int64(bizCtx.TenantID)
	if tenantID <= shared.PlatformTenantID {
		return 0, bizerr.NewCode(CodeTenantRequired)
	}
	return tenantID, nil
}

// ensureTenantScopedPlugin verifies the plugin exists and is tenant controllable.
func (s *serviceImpl) ensureTenantScopedPlugin(ctx context.Context, pluginID string) error {
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
		return bizerr.NewCode(CodePluginNotFound)
	}
	return nil
}

// tenantEnabled reads one tenant plugin enablement row.
func (s *serviceImpl) tenantEnabled(ctx context.Context, pluginID string, tenantID int64) (bool, error) {
	value, err := dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{
			PluginId: pluginID,
			TenantId: tenantID,
			StateKey: tenantEnablementStateKey,
		}).
		Value(dao.SysPluginState.Columns().Enabled)
	if err != nil {
		return false, err
	}
	return value != nil && !value.IsNil() && value.Bool(), nil
}

// pluginEntity converts a registry row and tenant flag into service output.
func pluginEntity(row *entity.SysPlugin, tenantEnabled bool) *Entity {
	enabled := 0
	if tenantEnabled {
		enabled = 1
	}
	return &Entity{
		Id:            row.PluginId,
		Name:          row.Name,
		Version:       row.Version,
		Type:          row.Type,
		Description:   row.Remark,
		Installed:     row.Installed,
		Enabled:       row.Status,
		ScopeNature:   row.ScopeNature,
		InstallMode:   row.InstallMode,
		TenantEnabled: enabled,
	}
}

// enablementStateValue converts the enablement flag to the persisted text value.
func enablementStateValue(enabled bool) string {
	if enabled {
		return tenantPluginEnabledValue
	}
	return tenantPluginDisabledValue
}
