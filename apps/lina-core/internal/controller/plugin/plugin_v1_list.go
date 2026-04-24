package plugin

import (
	"context"
	"strings"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginbridge"
)

// List scans plugins and returns current synchronized status list.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.pluginSvc.List(ctx, pluginsvc.ListInput{
		ID:        req.Id,
		Name:      req.Name,
		Type:      req.Type,
		Status:    req.Status,
		Installed: req.Installed,
	})
	if err != nil {
		return nil, err
	}

	tableComments := c.pluginSvc.ResolveDataTableComments(
		ctx,
		collectPluginDataAuthorizationTables(out.List),
	)
	managedCronJobsByPlugin := c.buildManagedCronJobMap(ctx, out.List)
	autoEnableManagedSet := buildAutoEnableManagedSet(c.configSvc.GetPluginAutoEnable(ctx))

	items := make([]*v1.PluginItem, 0, len(out.List))
	for _, item := range out.List {
		managedCronJobs := managedCronJobsByPlugin[item.Id]
		items = append(items, &v1.PluginItem{
			Id:                    item.Id,
			Name:                  item.Name,
			Version:               item.Version,
			Type:                  item.Type,
			Description:           item.Description,
			Installed:             item.Installed,
			InstalledAt:           item.InstalledAt,
			Enabled:               item.Enabled,
			AutoEnableManaged:     boolToInt(autoEnableManagedSet[strings.TrimSpace(item.Id)]),
			StatusKey:             item.StatusKey,
			UpdatedAt:             item.UpdatedAt,
			AuthorizationRequired: boolToInt(item.AuthorizationRequired),
			AuthorizationStatus:   string(item.AuthorizationStatus),
			RequestedHostServices: buildHostServicePermissionItems(
				item.RequestedHostServices,
				tableComments,
				managedCronJobs,
			),
			AuthorizedHostServices: buildHostServicePermissionItems(
				item.AuthorizedHostServices,
				tableComments,
				managedCronJobs,
			),
			DeclaredRoutes: buildPluginRouteReviewItems(
				item.Id,
				item.DeclaredRoutes,
			),
		})
	}

	return &v1.ListRes{List: items, Total: out.Total}, nil
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

// buildManagedCronJobMap loads plugin-owned cron declarations for plugins that
// expose the cron host service, so the review UI can present the discovered
// task summaries without blocking the list API on optional failures.
func (c *ControllerV1) buildManagedCronJobMap(
	ctx context.Context,
	items []*pluginsvc.PluginItem,
) map[string][]pluginsvc.ManagedCronJob {
	result := make(map[string][]pluginsvc.ManagedCronJob)
	for _, item := range items {
		if item == nil || strings.TrimSpace(item.Id) == "" {
			continue
		}
		if !pluginUsesCronHostService(item.RequestedHostServices) &&
			!pluginUsesCronHostService(item.AuthorizedHostServices) {
			continue
		}
		managedCronJobs, err := c.pluginSvc.ListManagedCronJobsByPlugin(ctx, item.Id)
		if err != nil {
			logger.Warningf(
				ctx,
				"load plugin managed cron jobs failed plugin=%s err=%v",
				item.Id,
				err,
			)
			continue
		}
		result[item.Id] = managedCronJobs
	}
	return result
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
	specs []*pluginbridge.HostServiceSpec,
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

// pluginUsesCronHostService reports whether the supplied host-service set
// contains the dedicated cron registration service.
func pluginUsesCronHostService(specs []*pluginbridge.HostServiceSpec) bool {
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		if strings.TrimSpace(spec.Service) == pluginbridge.HostServiceCron {
			return true
		}
	}
	return false
}

// boolToInt converts a boolean flag into the legacy integer representation used
// by list DTO fields.
func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
