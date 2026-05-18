// This file projects plugin dependency service DTOs into API response DTOs.
package plugin

import (
	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
)

// buildPluginDependencyCheckResult converts service dependency results to API DTOs.
func buildPluginDependencyCheckResult(in *pluginsvc.DependencyCheckResult) *v1.PluginDependencyCheckResult {
	if in == nil {
		return nil
	}
	return &v1.PluginDependencyCheckResult{
		TargetId: in.TargetID,
		Framework: v1.PluginDependencyFrameworkCheck{
			RequiredVersion: in.Framework.RequiredVersion,
			CurrentVersion:  in.Framework.CurrentVersion,
			Status:          in.Framework.Status,
		},
		Dependencies:          buildPluginDependencyItems(in.Dependencies),
		AutoInstallPlan:       buildPluginDependencyAutoInstallItems(in.AutoInstallPlan),
		AutoInstalled:         buildPluginDependencyAutoInstallItems(in.AutoInstalled),
		ManualInstallRequired: buildPluginDependencyItems(in.ManualInstallRequired),
		SoftUnsatisfied:       buildPluginDependencyItems(in.SoftUnsatisfied),
		Blockers:              buildPluginDependencyBlockers(in.Blockers),
		Cycle:                 cloneAPIStringSlice(in.Cycle),
		ReverseDependents:     buildPluginDependencyReverseDependents(in.ReverseDependents),
		ReverseBlockers:       buildPluginDependencyBlockers(in.ReverseBlockers),
	}
}

// buildPluginDependencyItems converts dependency edge DTOs.
func buildPluginDependencyItems(items []*pluginsvc.DependencyPluginCheck) []*v1.PluginDependencyItem {
	out := make([]*v1.PluginDependencyItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &v1.PluginDependencyItem{
			OwnerId:         item.OwnerID,
			DependencyId:    item.DependencyID,
			DependencyName:  item.DependencyName,
			RequiredVersion: item.RequiredVersion,
			CurrentVersion:  item.CurrentVersion,
			Required:        item.Required,
			InstallMode:     item.InstallMode,
			Installed:       item.Installed,
			Discovered:      item.Discovered,
			Status:          item.Status,
			Chain:           cloneAPIStringSlice(item.Chain),
		})
	}
	return out
}

// buildPluginDependencyAutoInstallItems converts auto-install plan DTOs.
func buildPluginDependencyAutoInstallItems(items []*pluginsvc.DependencyAutoInstallItem) []*v1.PluginDependencyAutoInstallItem {
	out := make([]*v1.PluginDependencyAutoInstallItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &v1.PluginDependencyAutoInstallItem{
			PluginId:   item.PluginID,
			Name:       item.Name,
			Version:    item.Version,
			RequiredBy: item.RequiredBy,
			Chain:      cloneAPIStringSlice(item.Chain),
		})
	}
	return out
}

// buildPluginDependencyBlockers converts hard dependency blocker DTOs.
func buildPluginDependencyBlockers(items []*pluginsvc.DependencyBlocker) []*v1.PluginDependencyBlocker {
	out := make([]*v1.PluginDependencyBlocker, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &v1.PluginDependencyBlocker{
			Code:            item.Code,
			PluginId:        item.PluginID,
			DependencyId:    item.DependencyID,
			RequiredVersion: item.RequiredVersion,
			CurrentVersion:  item.CurrentVersion,
			Chain:           cloneAPIStringSlice(item.Chain),
			Detail:          item.Detail,
		})
	}
	return out
}

// buildPluginDependencyReverseDependents converts downstream dependency DTOs.
func buildPluginDependencyReverseDependents(items []*pluginsvc.DependencyReverseDependent) []*v1.PluginDependencyReverseDependent {
	out := make([]*v1.PluginDependencyReverseDependent, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &v1.PluginDependencyReverseDependent{
			PluginId:        item.PluginID,
			Name:            item.Name,
			Version:         item.Version,
			RequiredVersion: item.RequiredVersion,
		})
	}
	return out
}

// cloneAPIStringSlice copies slices before exposing them through API DTOs.
func cloneAPIStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}
