// This file provides a lightweight filter runtime that maps plugin IDs to their
// installed-and-enabled status for OpenAPI route projection.

package openapi

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
)

// filterRuntime holds a snapshot of which plugins are currently enabled.
type filterRuntime struct {
	enabledByID map[string]bool
}

// buildFilterRuntime builds a filter runtime by querying the registry table for the
// installed and enabled status of every plugin in the manifest list.
func (s *serviceImpl) buildFilterRuntime(
	ctx context.Context,
	manifests []*catalog.Manifest,
) (*filterRuntime, error) {
	enabledByID := make(map[string]bool, len(manifests))
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
	}
	if len(enabledByID) == 0 {
		return &filterRuntime{enabledByID: enabledByID}, nil
	}

	registries, err := s.storeSvc.ListAllRegistries(ctx)
	if err != nil {
		return nil, err
	}
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		pluginID := strings.TrimSpace(registry.PluginId)
		if _, ok := enabledByID[pluginID]; !ok {
			continue
		}
		enabledByID[pluginID] = registry.Installed == plugintypes.InstalledYes &&
			registry.Status == plugintypes.StatusEnabled
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
