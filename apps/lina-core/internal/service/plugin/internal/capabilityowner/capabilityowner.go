// Package capabilityowner contains plugin-owned capability implementations
// that are wired into the plugin host service directory.
package capabilityowner

import (
	"context"
	"slices"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	pluginv1 "lina-core/api/plugin/v1"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/pluginconfig"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	capabilitytenantcap "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/statusflag"
)

const (
	tenantEnablementStateKey  = "__tenant_enabled__"
	tenantPluginEnabledValue  = "enabled"
	tenantPluginDisabledValue = "disabled"

	pluginRuntimeCacheDomain        cachecoord.Domain       = "plugin-runtime"
	tenantPluginRuntimeChangeReason cachecoord.ChangeReason = "tenant_plugin_enablement_changed"
)

// stateLookup defines host-owned plugin enablement lookups required by the
// plugin capability adapter without depending on the root plugin service
// package.
type stateLookup interface {
	// IsEnabled reports whether one plugin is enabled in the current scope.
	IsEnabled(ctx context.Context, pluginID string) bool
	// IsProviderEnabled reports whether one plugin may serve provider calls.
	IsProviderEnabled(ctx context.Context, pluginID string) bool
	// IsEnabledAuthoritative reports persisted plugin enablement bypassing local snapshots.
	IsEnabledAuthoritative(ctx context.Context, pluginID string) bool
}

// ScopedServicesFactory is implemented by host-owned capability directories
// that can project a plugin-bound service view.
type ScopedServicesFactory interface {
	// ForPlugin returns a capability service view bound to pluginID.
	ForPlugin(pluginID string) capability.Services
}

// CapabilityService exposes plugin governance services and scoped views.
type CapabilityService interface {
	capabilityplugincap.Service
	capabilitytenantcap.PluginService
	// ForPlugin returns a plugin-bound plugin-domain namespace.
	ForPlugin(pluginID string) capabilityplugincap.Service
}

// adapter implements the plugin-domain namespace and tenant enablement commands.
type pluginCapabilityAdapter struct {
	pluginID      string
	configFactory pluginconfig.Factory
	state         stateLookup
	lifecycle     capabilityplugincap.LifecycleService
	tenant        capabilitytenantcap.Service
	cacheCoordSvc cachecoord.Service
}

var (
	_ capabilityplugincap.Service         = (*pluginCapabilityAdapter)(nil)
	_ capabilityplugincap.RegistryService = (*pluginCapabilityAdapter)(nil)
	_ capabilityplugincap.StateService    = (*pluginCapabilityAdapter)(nil)
	_ capabilitytenantcap.PluginService   = (*pluginCapabilityAdapter)(nil)
)

// NewCapabilityAdapter creates the host-owned plugin governance capability adapter.
func NewCapabilityAdapter(
	configFactory pluginconfig.Factory,
	state stateLookup,
	lifecycle capabilityplugincap.LifecycleService,
	tenant capabilitytenantcap.Service,
	cacheCoordSvc cachecoord.Service,
) CapabilityService {
	return &pluginCapabilityAdapter{
		configFactory: configFactory,
		state:         state,
		lifecycle:     lifecycle,
		tenant:        tenant,
		cacheCoordSvc: cacheCoordSvc,
	}
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

// ServicesForPlugin returns a plugin-bound capability service set when the
// supplied host services support scoped binding; otherwise it returns the
// original services unchanged.
func ServicesForPlugin(services capability.Services, pluginID string) capability.Services {
	if services == nil {
		return nil
	}
	if scoped, ok := services.(ScopedServicesFactory); ok {
		return scoped.ForPlugin(strings.TrimSpace(pluginID))
	}
	return services
}

// Config returns the current plugin's static configuration reader.
func (a *pluginCapabilityAdapter) Config() capabilityplugincap.ConfigService {
	if a == nil || a.configFactory == nil || strings.TrimSpace(a.pluginID) == "" {
		return nil
	}
	return a.configFactory.ForPlugin(a.pluginID)
}

// Lifecycle returns plugin lifecycle governance operations.
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

// State returns plugin enablement lookup projections.
func (a *pluginCapabilityAdapter) State() capabilityplugincap.StateService {
	if a == nil {
		return nil
	}
	return a
}

// Current returns the projection for the current caller plugin.
func (a *pluginCapabilityAdapter) Current(ctx context.Context) (*capabilityplugincap.PluginInfo, error) {
	pluginID := strings.TrimSpace(a.pluginID)
	if pluginID == "" {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	result, err := a.BatchGet(ctx, []capabilityplugincap.PluginID{capabilityplugincap.PluginID(pluginID)})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if projection := result.Items[capabilityplugincap.PluginID(pluginID)]; projection != nil {
		return projection, nil
	}
	return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
}

// Get returns one visible plugin projection.
func (a *pluginCapabilityAdapter) Get(ctx context.Context, id capabilityplugincap.PluginID) (*capabilityplugincap.PluginInfo, error) {
	result, err := a.BatchGet(ctx, []capabilityplugincap.PluginID{id})
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

// BatchGet returns visible plugin projections and opaque missing IDs.
func (a *pluginCapabilityAdapter) BatchGet(ctx context.Context, ids []capabilityplugincap.PluginID) (*capmodel.BatchResult[*capabilityplugincap.PluginInfo, capabilityplugincap.PluginID], error) {
	result := &capmodel.BatchResult[*capabilityplugincap.PluginInfo, capabilityplugincap.PluginID]{
		Items:      make(map[capabilityplugincap.PluginID]*capabilityplugincap.PluginInfo, len(ids)),
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
		result.Items[requestID] = &capabilityplugincap.PluginInfo{
			ID:        capabilityplugincap.PluginID(row.PluginId),
			Version:   row.Version,
			Installed: row.Installed == statusflag.Installed.Int(),
			Enabled:   row.Status == statusflag.EnabledValue.Int(),
			Status:    row.CurrentState,
		}
	}
	for _, id := range ids {
		if _, ok := result.Items[id]; !ok && !slices.Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// List returns bounded plugin governance projections.
func (a *pluginCapabilityAdapter) List(ctx context.Context, input capabilityplugincap.ListInput) (*capmodel.PageResult[*capabilityplugincap.PluginInfo], error) {
	pageNum, pageSize := input.Page.Normalize()
	if pageSize > capabilityplugincap.MaxPluginSearchPageSize {
		pageSize = capabilityplugincap.MaxPluginSearchPageSize
	}
	cols := dao.SysPlugin.Columns()
	model := dao.SysPlugin.Ctx(ctx)
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		model = model.Where(model.Builder().
			WhereLike(cols.PluginId, like).
			WhereOrLike(cols.Name, like).
			WhereOrLike(cols.Remark, like))
	}
	if pluginID := strings.TrimSpace(input.PluginID); pluginID != "" {
		model = model.WhereLike(cols.PluginId, "%"+pluginID+"%")
	}
	if name := strings.TrimSpace(input.Name); name != "" {
		model = model.WhereLike(cols.Name, "%"+name+"%")
	}
	if pluginType := strings.TrimSpace(string(input.Type)); pluginType != "" {
		model = model.Where(cols.Type, pluginType)
	}
	if input.Enabled != nil {
		status := 0
		if *input.Enabled {
			status = statusflag.EnabledValue.Int()
		}
		model = model.Where(cols.Status, status)
	}
	total, err := model.Clone().Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.SysPlugin, 0, pageSize)
	if err = model.Clone().
		Fields(cols.PluginId, cols.Version, cols.Installed, cols.Status, cols.CurrentState).
		OrderAsc(cols.PluginId).
		Page(pageNum, pageSize).
		Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*capabilityplugincap.PluginInfo, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, pluginProjection(row))
	}
	return &capmodel.PageResult[*capabilityplugincap.PluginInfo]{Items: items, Total: total}, nil
}

// ListTenantPlugins returns tenant-controllable plugins with current tenant enablement.
func (a *pluginCapabilityAdapter) ListTenantPlugins(ctx context.Context, input capabilityplugincap.TenantListInput) (*capmodel.PageResult[*capabilityplugincap.TenantPluginInfo], error) {
	tenantID := a.currentTenantID(ctx)
	if tenantID <= datascope.PlatformTenantID {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	pageNum, pageSize := input.Page.Normalize()
	if pageSize > capabilityplugincap.MaxPluginSearchPageSize {
		pageSize = capabilityplugincap.MaxPluginSearchPageSize
	}
	cols := dao.SysPlugin.Columns()
	model := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{
			Installed:   statusflag.Installed.Int(),
			Status:      statusflag.EnabledValue.Int(),
			ScopeNature: string(pluginv1.ScopeNatureTenantAware),
			InstallMode: string(pluginv1.InstallModeTenantScoped),
		})
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		model = model.Where(model.Builder().
			WhereLike(cols.PluginId, like).
			WhereOrLike(cols.Name, like).
			WhereOrLike(cols.Remark, like))
	}
	if pluginType := strings.TrimSpace(string(input.Type)); pluginType != "" {
		model = model.Where(cols.Type, pluginType)
	}
	total, err := model.Clone().Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.SysPlugin, 0)
	if err = model.Clone().
		OrderAsc(cols.PluginId).
		Page(pageNum, pageSize).
		Scan(&rows); err != nil {
		return nil, err
	}
	states, err := a.tenantEnabledStates(ctx, tenantID, pluginIDsFromRows(rows))
	if err != nil {
		return nil, err
	}
	items := make([]*capabilityplugincap.TenantPluginInfo, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		tenantEnabled := states[row.PluginId]
		if input.TenantEnabled != nil && tenantEnabled != *input.TenantEnabled {
			continue
		}
		items = append(items, tenantPluginProjection(row, tenantEnabled))
	}
	if input.TenantEnabled != nil {
		total = len(items)
	}
	return &capmodel.PageResult[*capabilityplugincap.TenantPluginInfo]{Items: items, Total: total}, nil
}

// IsEnabled reports whether one plugin is enabled in the current scope.
func (a *pluginCapabilityAdapter) IsEnabled(ctx context.Context, pluginID capabilityplugincap.PluginID) (bool, error) {
	projection, err := a.Get(ctx, pluginID)
	if err != nil {
		return false, err
	}
	return projection.Enabled, nil
}

// IsProviderEnabled reports whether one plugin may serve provider calls.
func (a *pluginCapabilityAdapter) IsProviderEnabled(ctx context.Context, pluginID capabilityplugincap.PluginID) (bool, error) {
	if a == nil || a.state == nil {
		return false, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "plugin-state"))
	}
	return a.state.IsProviderEnabled(ctx, string(pluginID)), nil
}

// IsEnabledAuthoritative reports persisted plugin enablement bypassing local snapshots.
func (a *pluginCapabilityAdapter) IsEnabledAuthoritative(ctx context.Context, pluginID capabilityplugincap.PluginID) (bool, error) {
	if a == nil || a.state == nil {
		return false, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "plugin-state"))
	}
	return a.state.IsEnabledAuthoritative(ctx, string(pluginID)), nil
}

// SetTenantPluginEnabled changes tenant plugin enablement after target checks.
func (a *pluginCapabilityAdapter) SetTenantPluginEnabled(ctx context.Context, id capabilityplugincap.PluginID, enabled bool) error {
	tenantID := a.currentTenantID(ctx)
	if tenantID <= datascope.PlatformTenantID {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	pluginID := strings.TrimSpace(string(id))
	if pluginID == "" {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if err := a.ensureTenantScopedPlugin(ctx, pluginID); err != nil {
		return err
	}
	return a.setTenantPluginEnabled(ctx, tenantID, pluginID, enabled, false)
}

// ProvisionTenantPluginDefaults creates missing default tenant plugin enablement rows.
func (a *pluginCapabilityAdapter) ProvisionTenantPluginDefaults(ctx context.Context, tenantIDValue capmodel.DomainID) error {
	tenantID, err := capabilitytenantcap.ParseTenantID(tenantIDValue)
	if err != nil || tenantID <= datascope.PlatformTenantID {
		return nil
	}
	rows := make([]*entity.SysPlugin, 0)
	if err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{
			Installed:               statusflag.Installed.Int(),
			Status:                  statusflag.EnabledValue.Int(),
			ScopeNature:             string(pluginv1.ScopeNatureTenantAware),
			InstallMode:             string(pluginv1.InstallModeTenantScoped),
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

func (a *pluginCapabilityAdapter) currentTenantID(ctx context.Context) int {
	if a == nil || a.tenant == nil || a.tenant.Context() == nil {
		return datascope.PlatformTenantID
	}
	return int(a.tenant.Context().Current(ctx))
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
			Installed:   statusflag.Installed.Int(),
			Status:      statusflag.EnabledValue.Int(),
			ScopeNature: string(pluginv1.ScopeNatureTenantAware),
			InstallMode: string(pluginv1.InstallModeTenantScoped),
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
	if a == nil || a.cacheCoordSvc == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "cachecoord"))
	}
	identity := do.SysPluginState{
		PluginId: pluginID,
		TenantId: tenantID,
		StateKey: tenantEnablementStateKey,
	}
	changed := false
	if err := dao.SysPluginState.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		result, insertErr := dao.SysPluginState.Ctx(ctx).Data(do.SysPluginState{
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
			changed = true
			return nil
		}
		_, updateErr := dao.SysPluginState.Ctx(ctx).
			Where(identity).
			Data(do.SysPluginState{
				StateValue: pluginEnablementStateValue(enabled),
				Enabled:    enabled,
			}).
			Update()
		if updateErr != nil {
			return updateErr
		}
		changed = true
		return nil
	}); err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return a.bumpPluginRuntimeCacheRevision(ctx)
}

// bumpPluginRuntimeCacheRevision advances the shared plugin-runtime revision.
func (a *pluginCapabilityAdapter) bumpPluginRuntimeCacheRevision(ctx context.Context) error {
	_, err := a.cacheCoordSvc.MarkChanged(
		ctx,
		pluginRuntimeCacheDomain,
		cachecoord.ScopeGlobal,
		tenantPluginRuntimeChangeReason,
	)
	return err
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

// pluginProjection converts one host plugin registry row into a stable projection.
func pluginProjection(row *entity.SysPlugin) *capabilityplugincap.PluginInfo {
	return &capabilityplugincap.PluginInfo{
		ID:        capabilityplugincap.PluginID(row.PluginId),
		Version:   row.Version,
		Installed: row.Installed == statusflag.Installed.Int(),
		Enabled:   row.Status == statusflag.EnabledValue.Int(),
		Status:    row.CurrentState,
	}
}

// tenantPluginProjection converts a host plugin registry row into a stable projection.
func tenantPluginProjection(row *entity.SysPlugin, tenantEnabled bool) *capabilityplugincap.TenantPluginInfo {
	return &capabilityplugincap.TenantPluginInfo{
		ID:            capabilityplugincap.PluginID(row.PluginId),
		Name:          row.Name,
		Version:       row.Version,
		Type:          pluginv1.PluginType(row.Type),
		Description:   row.Remark,
		Installed:     row.Installed == statusflag.Installed.Int(),
		Enabled:       row.Status == statusflag.EnabledValue.Int(),
		ScopeNature:   pluginv1.ScopeNature(row.ScopeNature),
		InstallMode:   pluginv1.InstallMode(row.InstallMode),
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
