package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
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

	items := make([]*v1.PluginItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.PluginItem{
			Id:                    item.Id,
			Name:                  item.Name,
			Version:               item.Version,
			Type:                  item.Type,
			Description:           item.Description,
			Installed:             item.Installed,
			InstalledAt:           item.InstalledAt,
			Enabled:               item.Enabled,
			StatusKey:             item.StatusKey,
			UpdatedAt:             item.UpdatedAt,
			AuthorizationRequired: boolToInt(item.AuthorizationRequired),
			AuthorizationStatus:   string(item.AuthorizationStatus),
			RequestedHostServices: buildHostServicePermissionItems(
				item.RequestedHostServices,
				tableComments,
			),
			AuthorizedHostServices: buildHostServicePermissionItems(
				item.AuthorizedHostServices,
				tableComments,
			),
		})
	}

	return &v1.ListRes{List: items, Total: out.Total}, nil
}

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

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
