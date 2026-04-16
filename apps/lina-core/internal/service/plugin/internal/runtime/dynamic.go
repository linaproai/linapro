// This file exposes public runtime-state projections consumed by plugin-aware
// frontend shells that need minimal installation and enablement state.

package runtime

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
)

// RuntimeStateListOutput defines output for public runtime state queries.
type RuntimeStateListOutput struct {
	// List contains public plugin runtime states.
	List []*PluginDynamicStateItem
}

// PluginDynamicStateItem represents public runtime state of one plugin.
type PluginDynamicStateItem struct {
	// Id is the stable plugin identifier.
	Id string
	// Installed reports whether the plugin is installed or integrated.
	Installed int
	// Enabled reports whether the plugin is currently enabled.
	Enabled int
	// Version is the currently active plugin version.
	Version string
	// Generation is the current active plugin generation on the host.
	Generation int64
	// StatusKey is the host config key used by the public shell.
	StatusKey string
}

// ListRuntimeStates returns public plugin runtime states for shell slot rendering.
func (s *serviceImpl) ListRuntimeStates(ctx context.Context) (*RuntimeStateListOutput, error) {
	registries, err := s.catalogSvc.ListAllRegistries(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*PluginDynamicStateItem, 0, len(registries))
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		pluginID := strings.TrimSpace(registry.PluginId)
		if pluginID == "" {
			continue
		}

		installed := registry.Installed
		enabled := registry.Status
		if catalog.NormalizeType(registry.Type) == catalog.TypeDynamic {
			exists, _, err := s.hasArtifactStorageFile(ctx, pluginID)
			if err != nil {
				return nil, err
			}
			if !exists {
				installed = catalog.InstalledNo
				enabled = catalog.StatusDisabled
			}
		}

		generation := registry.Generation
		if generation <= 0 {
			generation = 1
		}

		items = append(items, &PluginDynamicStateItem{
			Id:         pluginID,
			Installed:  installed,
			Enabled:    enabled,
			Version:    registry.Version,
			Generation: generation,
			StatusKey:  s.catalogSvc.BuildPluginStatusKey(pluginID),
		})
	}
	return &RuntimeStateListOutput{List: items}, nil
}
