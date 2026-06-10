// This file implements plugin list projection for the management API.

package plugin

import (
	"context"
	"strings"

	v1 "lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/statusflag"
)

// List scans plugins and returns current synchronized status list.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.pluginSvc.List(ctx, pluginsvc.ListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		ID:        req.Id,
		Name:      req.Name,
		Type:      string(req.Type),
		Status:    enabledPtrToInt(req.Status),
		Installed: installationPtrToInt(req.Installed),
	})
	if err != nil {
		return nil, err
	}

	autoEnableManagedSet := buildAutoEnableManagedSet(c.configSvc.GetPluginAutoEnable(ctx))

	items := make([]*v1.PluginListItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, c.buildPluginListItemResponse(
			item,
			autoEnableManagedSet[strings.TrimSpace(item.Id)],
		))
	}

	return &v1.ListRes{List: items, Total: out.Total}, nil
}

// buildPluginListItemResponse maps the service plugin summary projection into
// the public list DTO. Detail-only governance payloads are intentionally absent.
func (c *ControllerV1) buildPluginListItemResponse(
	item *pluginsvc.PluginItem,
	autoEnableManaged bool,
) *v1.PluginListItem {
	if item == nil {
		return nil
	}
	return &v1.PluginListItem{
		Id:                      item.Id,
		Name:                    item.Name,
		Version:                 item.Version,
		RuntimeState:            v1.RuntimeState(item.RuntimeState.String()),
		EffectiveVersion:        item.EffectiveVersion,
		DiscoveredVersion:       item.DiscoveredVersion,
		UpgradeAvailable:        item.UpgradeAvailable,
		AbnormalReason:          v1.RuntimeAbnormalReason(item.AbnormalReason.String()),
		LastUpgradeFailure:      buildPluginUpgradeFailureItem(item.LastUpgradeFailure),
		Type:                    v1.PluginType(item.Type),
		Description:             item.Description,
		Installed:               statusflag.Installation(item.Installed),
		InstalledAt:             item.InstalledAt,
		Enabled:                 statusflag.Enabled(item.Enabled),
		AutoEnableManaged:       boolToYesNo(autoEnableManaged),
		AutoEnableForNewTenants: item.AutoEnableForNewTenants,
		SupportsMultiTenant:     item.SupportsMultiTenant,
		ScopeNature:             v1.ScopeNature(item.ScopeNature),
		InstallMode:             v1.InstallMode(item.InstallMode),
		StatusKey:               item.StatusKey,
		UpdatedAt:               item.UpdatedAt,
		AuthorizationRequired:   boolToYesNo(item.AuthorizationRequired),
		AuthorizationStatus:     v1.AuthorizationStatus(item.AuthorizationStatus),
		HasMockData:             boolToYesNo(item.HasMockData),
	}
}

// buildPluginItemResponse maps the service plugin detail projection into the
// public management DTO used by detail and action review endpoints.
func (c *ControllerV1) buildPluginItemResponse(
	item *pluginsvc.PluginItem,
	tableComments map[string]string,
	autoEnableManaged bool,
) *v1.PluginItem {
	if item == nil {
		return nil
	}
	return &v1.PluginItem{
		Id:                      item.Id,
		Name:                    item.Name,
		Version:                 item.Version,
		RuntimeState:            v1.RuntimeState(item.RuntimeState.String()),
		EffectiveVersion:        item.EffectiveVersion,
		DiscoveredVersion:       item.DiscoveredVersion,
		UpgradeAvailable:        item.UpgradeAvailable,
		AbnormalReason:          v1.RuntimeAbnormalReason(item.AbnormalReason.String()),
		LastUpgradeFailure:      buildPluginUpgradeFailureItem(item.LastUpgradeFailure),
		Type:                    v1.PluginType(item.Type),
		Description:             item.Description,
		Installed:               statusflag.Installation(item.Installed),
		InstalledAt:             item.InstalledAt,
		Enabled:                 statusflag.Enabled(item.Enabled),
		AutoEnableManaged:       boolToYesNo(autoEnableManaged),
		AutoEnableForNewTenants: item.AutoEnableForNewTenants,
		SupportsMultiTenant:     item.SupportsMultiTenant,
		ScopeNature:             v1.ScopeNature(item.ScopeNature),
		InstallMode:             v1.InstallMode(item.InstallMode),
		StatusKey:               item.StatusKey,
		UpdatedAt:               item.UpdatedAt,
		AuthorizationRequired:   boolToYesNo(item.AuthorizationRequired),
		AuthorizationStatus:     v1.AuthorizationStatus(item.AuthorizationStatus),
		DependencyCheck:         buildPluginDependencyCheckResult(item.DependencyCheck),
		RequestedHostServices: buildHostServicePermissionItems(
			item.RequestedHostServices,
			tableComments,
		),
		AuthorizedHostServices: buildHostServicePermissionItems(
			item.AuthorizedHostServices,
			tableComments,
		),
		DeclaredRoutes: buildPluginRouteReviewItems(
			item.Id,
			item.DeclaredRoutes,
		),
		HasMockData: boolToYesNo(item.HasMockData),
	}
}

// buildAutoEnableManagedSet converts the normalized plugin.autoEnable list into
// one lookup map that the controller can reuse while projecting list rows.
func buildAutoEnableManagedSet(pluginIDs []string) map[string]bool {
	managedSet := make(map[string]bool, len(pluginIDs))
	for _, pluginID := range pluginIDs {
		normalizedPluginID := strings.TrimSpace(pluginID)
		if normalizedPluginID == "" {
			continue
		}
		managedSet[normalizedPluginID] = true
	}
	return managedSet
}

// collectPluginDataAuthorizationTables gathers the unique host data-table names
// referenced by requested and authorized plugin host-service specs.
func collectPluginDataAuthorizationTables(items []*pluginsvc.PluginItem) []string {
	tableSet := make(map[string]struct{})
	tables := make([]string, 0)
	for _, item := range items {
		if item == nil {
			continue
		}
		collectHostServiceTables(tableSet, &tables, item.RequestedHostServices)
		collectHostServiceTables(tableSet, &tables, item.AuthorizedHostServices)
	}
	return tables
}

// collectHostServiceTables appends previously unseen table names from the
// supplied host-service specs into the target slice.
func collectHostServiceTables(
	tableSet map[string]struct{},
	tables *[]string,
	specs []*protocol.HostServiceSpec,
) {
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		for _, table := range spec.Tables {
			if _, ok := tableSet[table]; ok {
				continue
			}
			tableSet[table] = struct{}{}
			*tables = append(*tables, table)
		}
	}
}

// buildPluginUpgradeFailureItem maps the service runtime-upgrade failure
// projection into the public plugin list DTO.
func buildPluginUpgradeFailureItem(
	failure *pluginsvc.RuntimeUpgradeFailure,
) *v1.PluginUpgradeFailureItem {
	if failure == nil {
		return nil
	}
	return &v1.PluginUpgradeFailureItem{
		Phase:          v1.RuntimeFailurePhase(failure.Phase.String()),
		ErrorCode:      failure.ErrorCode,
		MessageKey:     failure.MessageKey,
		ReleaseId:      failure.ReleaseID,
		ReleaseVersion: failure.ReleaseVersion,
		Detail:         failure.Detail,
	}
}
