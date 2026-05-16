// This file implements the plugin runtime-upgrade preview management API.

package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
)

// UpgradePreview returns a side-effect-free runtime upgrade preview.
func (c *ControllerV1) UpgradePreview(
	ctx context.Context,
	req *v1.UpgradePreviewReq,
) (res *v1.UpgradePreviewRes, err error) {
	preview, err := c.pluginSvc.PreviewRuntimeUpgrade(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	tableComments := c.pluginSvc.ResolveDataTableComments(
		ctx,
		collectUpgradePreviewDataAuthorizationTables(preview),
	)
	return buildUpgradePreviewResponse(tableComments, preview), nil
}

// buildUpgradePreviewResponse maps service preview output into the public DTO.
func buildUpgradePreviewResponse(
	tableComments map[string]string,
	preview *pluginsvc.RuntimeUpgradePreview,
) *v1.UpgradePreviewRes {
	if preview == nil {
		return nil
	}
	return &v1.UpgradePreviewRes{
		PluginId:          preview.PluginID,
		RuntimeState:      preview.RuntimeState.String(),
		EffectiveVersion:  preview.EffectiveVersion,
		DiscoveredVersion: preview.DiscoveredVersion,
		FromManifest:      buildPluginManifestSnapshotItem(tableComments, preview.FromManifest),
		ToManifest:        buildPluginManifestSnapshotItem(tableComments, preview.ToManifest),
		DependencyCheck:   buildPluginDependencyCheckResult(preview.DependencyCheck),
		SQLSummary: v1.PluginUpgradeSQLSummary{
			InstallSQLCount:      preview.SQLSummary.InstallSQLCount,
			UninstallSQLCount:    preview.SQLSummary.UninstallSQLCount,
			MockSQLCount:         preview.SQLSummary.MockSQLCount,
			RuntimeSQLAssetCount: preview.SQLSummary.RuntimeSQLAssetCount,
		},
		HostServicesDiff: buildPluginUpgradeHostServicesDiff(preview.HostServicesDiff),
		RiskHints:        cloneAPIStringSlice(preview.RiskHints),
	}
}

// buildPluginManifestSnapshotItem maps one manifest snapshot into preview DTO.
func buildPluginManifestSnapshotItem(
	tableComments map[string]string,
	snapshot *pluginsvc.RuntimeUpgradeManifestSnapshot,
) *v1.PluginManifestSnapshotItem {
	if snapshot == nil {
		return nil
	}
	return &v1.PluginManifestSnapshotItem{
		Id:                        snapshot.ID,
		Name:                      snapshot.Name,
		Version:                   snapshot.Version,
		Type:                      snapshot.Type,
		ScopeNature:               snapshot.ScopeNature,
		SupportsMultiTenant:       snapshot.SupportsMultiTenant,
		DefaultInstallMode:        snapshot.DefaultInstallMode,
		Description:               snapshot.Description,
		RuntimeKind:               snapshot.RuntimeKind,
		RuntimeAbiVersion:         snapshot.RuntimeABIVersion,
		ManifestDeclared:          snapshot.ManifestDeclared,
		InstallSQLCount:           snapshot.InstallSQLCount,
		UninstallSQLCount:         snapshot.UninstallSQLCount,
		MockSQLCount:              snapshot.MockSQLCount,
		FrontendPageCount:         snapshot.FrontendPageCount,
		FrontendSlotCount:         snapshot.FrontendSlotCount,
		MenuCount:                 snapshot.MenuCount,
		BackendHookCount:          snapshot.BackendHookCount,
		ResourceSpecCount:         snapshot.ResourceSpecCount,
		RouteCount:                snapshot.RouteCount,
		RouteExecutionEnabled:     snapshot.RouteExecutionEnabled,
		RouteRequestCodec:         snapshot.RouteRequestCodec,
		RouteResponseCodec:        snapshot.RouteResponseCodec,
		RuntimeFrontendAssetCount: snapshot.RuntimeFrontendAssetCount,
		RuntimeSQLAssetCount:      snapshot.RuntimeSQLAssetCount,
		HostServiceAuthRequired:   snapshot.HostServiceAuthRequired,
		HostServiceAuthConfirmed:  snapshot.HostServiceAuthConfirmed,
		RequestedHostServices: buildHostServicePermissionItems(
			snapshot.RequestedHostServices,
			tableComments,
			nil,
		),
		AuthorizedHostServices: buildHostServicePermissionItems(
			snapshot.AuthorizedHostServices,
			tableComments,
			nil,
		),
	}
}

// buildPluginUpgradeHostServicesDiff maps service hostServices diff output to API DTO.
func buildPluginUpgradeHostServicesDiff(
	diff pluginsvc.RuntimeUpgradeHostServicesDiff,
) v1.PluginUpgradeHostServicesDiff {
	return v1.PluginUpgradeHostServicesDiff{
		Added:                 buildPluginUpgradeHostServiceChanges(diff.Added),
		Removed:               buildPluginUpgradeHostServiceChanges(diff.Removed),
		Changed:               buildPluginUpgradeHostServiceChanges(diff.Changed),
		AuthorizationRequired: diff.AuthorizationRequired,
		AuthorizationChanged:  diff.AuthorizationChanged,
	}
}

// buildPluginUpgradeHostServiceChanges maps host-service change summaries.
func buildPluginUpgradeHostServiceChanges(
	items []*pluginsvc.RuntimeUpgradeHostServiceChange,
) []*v1.PluginUpgradeHostServiceChange {
	out := make([]*v1.PluginUpgradeHostServiceChange, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &v1.PluginUpgradeHostServiceChange{
			Service:           item.Service,
			FromMethods:       cloneAPIStringSlice(item.FromMethods),
			ToMethods:         cloneAPIStringSlice(item.ToMethods),
			FromResourceCount: item.FromResourceCount,
			ToResourceCount:   item.ToResourceCount,
			FromTables:        cloneAPIStringSlice(item.FromTables),
			ToTables:          cloneAPIStringSlice(item.ToTables),
			FromPaths:         cloneAPIStringSlice(item.FromPaths),
			ToPaths:           cloneAPIStringSlice(item.ToPaths),
		})
	}
	return out
}

// collectUpgradePreviewDataAuthorizationTables gathers host data table names
// referenced by manifest snapshots in the preview.
func collectUpgradePreviewDataAuthorizationTables(preview *pluginsvc.RuntimeUpgradePreview) []string {
	if preview == nil {
		return nil
	}
	tableSet := make(map[string]struct{})
	tables := make([]string, 0)
	if preview.FromManifest != nil {
		collectHostServiceTables(tableSet, &tables, preview.FromManifest.RequestedHostServices)
		collectHostServiceTables(tableSet, &tables, preview.FromManifest.AuthorizedHostServices)
	}
	if preview.ToManifest != nil {
		collectHostServiceTables(tableSet, &tables, preview.ToManifest.RequestedHostServices)
		collectHostServiceTables(tableSet, &tables, preview.ToManifest.AuthorizedHostServices)
	}
	return tables
}
