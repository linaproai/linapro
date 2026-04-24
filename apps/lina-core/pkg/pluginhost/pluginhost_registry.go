// This file stores the in-memory registry of compile-time source plugins that
// are linked into the host binary during build time.

package pluginhost

import "sync"

// In-memory source-plugin registry shared by build-linked plugins.
var (
	sourcePluginRegistryMu sync.RWMutex
	sourcePluginRegistry   = make(map[string]SourcePluginDefinition)
	sourcePluginListeners  []func()
)

// RegisterSourcePlugin registers one compile-time source plugin into the host registry.
func RegisterSourcePlugin(plugin SourcePlugin) {
	if plugin == nil {
		panic("pluginhost: source plugin is nil")
	}
	definition, ok := plugin.(SourcePluginDefinition)
	if !ok {
		panic("pluginhost: source plugin does not implement SourcePluginDefinition")
	}
	if definition.ID() == "" {
		panic("pluginhost: source plugin id is empty")
	}

	sourcePluginRegistryMu.Lock()
	if _, exists := sourcePluginRegistry[definition.ID()]; exists {
		sourcePluginRegistryMu.Unlock()
		panic("pluginhost: duplicate source plugin registration: " + definition.ID())
	}
	sourcePluginRegistry[definition.ID()] = definition
	listeners := append([]func(){}, sourcePluginListeners...)
	sourcePluginRegistryMu.Unlock()
	notifySourcePluginListeners(listeners)
}

// RegisterSourcePluginRegistryListener adds one listener that is invoked after
// compile-time source-plugin registrations change the registry content.
func RegisterSourcePluginRegistryListener(listener func()) {
	if listener == nil {
		return
	}

	sourcePluginRegistryMu.Lock()
	defer sourcePluginRegistryMu.Unlock()

	sourcePluginListeners = append(sourcePluginListeners, listener)
}

// GetSourcePlugin returns one registered compile-time source plugin by id.
func GetSourcePlugin(id string) (SourcePluginDefinition, bool) {
	sourcePluginRegistryMu.RLock()
	defer sourcePluginRegistryMu.RUnlock()

	plugin, ok := sourcePluginRegistry[id]
	return plugin, ok
}

// ListSourcePlugins returns all registered compile-time source plugins.
func ListSourcePlugins() []SourcePluginDefinition {
	sourcePluginRegistryMu.RLock()
	defer sourcePluginRegistryMu.RUnlock()

	items := make([]SourcePluginDefinition, 0, len(sourcePluginRegistry))
	for _, plugin := range sourcePluginRegistry {
		if plugin == nil {
			continue
		}
		items = append(items, plugin)
	}
	return items
}

// notifySourcePluginListeners invokes the current registry listeners outside
// the registry lock so callbacks may safely query other pluginhost APIs.
func notifySourcePluginListeners(listeners []func()) {
	for _, listener := range listeners {
		if listener != nil {
			listener()
		}
	}
}
