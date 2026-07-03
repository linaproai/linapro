// This file tests plugin-list controller helpers that project startup
// auto-enable configuration into management-list view fields.

package plugin

import (
	"testing"

	v1 "lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
)

// TestBuildAutoEnableManagedSetNormalizesPluginIDs verifies startup-managed
// plugin IDs are trimmed and de-duplicated before list projection uses them.
func TestBuildAutoEnableManagedSetNormalizesPluginIDs(t *testing.T) {
	managedSet := buildAutoEnableManagedSet([]string{
		" linapro-demo-source ",
		"linapro-demo-source",
		"",
		"linapro-monitor-server",
	})

	if len(managedSet) != 2 {
		t.Fatalf("expected 2 unique managed plugin IDs, got %d", len(managedSet))
	}
	if !managedSet["linapro-demo-source"] {
		t.Fatal("expected linapro-demo-source to be marked as managed")
	}
	if !managedSet["linapro-monitor-server"] {
		t.Fatal("expected linapro-monitor-server to be marked as managed")
	}
	if managedSet[""] {
		t.Fatal("expected blank plugin ID to be ignored")
	}
}

// TestBuildPluginListItemResponseIncludesDistribution verifies API projection
// exposes plugin distribution governance metadata.
func TestBuildPluginListItemResponseIncludesDistribution(t *testing.T) {
	controller := &ControllerV1{}
	pluginItem := &pluginsvc.PluginItem{}
	pluginItem.Id = "plugin-dev-controller-builtin"
	pluginItem.Name = "Controller Builtin"
	pluginItem.Type = string(v1.PluginTypeSource)
	pluginItem.Distribution = string(v1.PluginDistributionBuiltin)

	item := controller.buildPluginListItemResponse(pluginItem, false)
	if item == nil {
		t.Fatal("expected plugin list item response")
	}
	if item.Distribution != v1.PluginDistributionBuiltin {
		t.Fatalf("expected builtin distribution in API projection, got %q", item.Distribution)
	}
}
