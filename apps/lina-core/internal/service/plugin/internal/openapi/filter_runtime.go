// This file provides a lightweight filter runtime that maps plugin IDs to their
// installed-and-enabled status for OpenAPI route projection.

package openapi

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

// filterRuntime holds a snapshot of which plugins are currently enabled.
type filterRuntime struct {
	enabledByID map[string]bool
}

// buildFilterRuntime builds a filter runtime by querying the registry table for the
// installed and enabled status of every plugin in the manifest list.
func buildFilterRuntime(
	ctx context.Context,
	manifests []*catalog.Manifest,
) (*filterRuntime, error) {
	enabledByID := make(map[string]bool, len(manifests))
	pluginIDs := make([]string, 0, len(manifests))
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		pluginID := strings.TrimSpace(manifest.ID)
		if pluginID == "" {
			continue
		}
		if _, ok := enabledByID[pluginID]; ok {
			continue
		}
		enabledByID[pluginID] = false
		pluginIDs = append(pluginIDs, pluginID)
	}
	if len(pluginIDs) == 0 {
		return &filterRuntime{enabledByID: enabledByID}, nil
	}

	var registries []*entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		WhereIn(dao.SysPlugin.Columns().PluginId, pluginIDs).
		Scan(&registries)
	if err != nil {
		return nil, err
	}

	for _, registry := range registries {
		if registry == nil {
			continue
		}
		pluginID := strings.TrimSpace(registry.PluginId)
		enabledByID[pluginID] = registry.Installed == catalog.InstalledYes &&
			registry.Status == catalog.StatusEnabled
	}
	return &filterRuntime{enabledByID: enabledByID}, nil
}

// isEnabled reports whether the plugin with the given ID is currently installed and enabled.
func (r *filterRuntime) isEnabled(pluginID string) bool {
	if r == nil {
		return false
	}
	return r.enabledByID[strings.TrimSpace(pluginID)]
}
