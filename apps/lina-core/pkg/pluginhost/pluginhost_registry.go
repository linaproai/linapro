// This file stores the in-memory registry of compile-time source plugins that
// are linked into the host binary during build time.

package pluginhost

import "sync"

var (
	sourcePluginRegistryMu sync.RWMutex
	sourcePluginRegistry   = make(map[string]*SourcePlugin)
)

// RegisterSourcePlugin registers one compile-time source plugin into the host registry.
func RegisterSourcePlugin(plugin *SourcePlugin) {
	if plugin == nil {
		panic("pluginhost: source plugin is nil")
	}
	if plugin.ID == "" {
		panic("pluginhost: source plugin id is empty")
	}

	sourcePluginRegistryMu.Lock()
	defer sourcePluginRegistryMu.Unlock()

	if _, ok := sourcePluginRegistry[plugin.ID]; ok {
		panic("pluginhost: duplicate source plugin registration: " + plugin.ID)
	}
	sourcePluginRegistry[plugin.ID] = plugin
}

// GetSourcePlugin returns one registered compile-time source plugin by id.
func GetSourcePlugin(id string) (*SourcePlugin, bool) {
	sourcePluginRegistryMu.RLock()
	defer sourcePluginRegistryMu.RUnlock()

	plugin, ok := sourcePluginRegistry[id]
	return plugin, ok
}

// ListSourcePlugins returns all registered compile-time source plugins.
func ListSourcePlugins() []*SourcePlugin {
	sourcePluginRegistryMu.RLock()
	defer sourcePluginRegistryMu.RUnlock()

	items := make([]*SourcePlugin, 0, len(sourcePluginRegistry))
	for _, plugin := range sourcePluginRegistry {
		if plugin == nil {
			continue
		}
		items = append(items, plugin)
	}
	return items
}
