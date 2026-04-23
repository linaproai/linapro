// This file tests that multiple integration service instances share the same
// in-memory source-plugin enablement and route-binding state inside one host
// process.

package integration

import (
	"testing"

	"lina-core/pkg/pluginhost"
)

// TestSharedStateCrossInstance verifies route guards and route-binding
// projections stay aligned even when different callers hold different
// integration service instances.
func TestSharedStateCrossInstance(t *testing.T) {
	shared := &sharedState{
		sourceRouteBindings: make(map[string][]pluginhost.SourceRouteBinding),
		enabledSnapshot:     make(map[string]bool),
	}
	first := &serviceImpl{sharedState: shared}
	second := &serviceImpl{sharedState: shared}

	first.SetPluginEnabledState("plugin-demo", true)

	firstChecker := first.buildBackgroundEnabledChecker()
	secondChecker := second.buildBackgroundEnabledChecker()
	if !firstChecker("plugin-demo") {
		t.Fatal("expected first instance to read enabled snapshot")
	}
	if !secondChecker("plugin-demo") {
		t.Fatal("expected second instance to share enabled snapshot updates")
	}

	first.setSourceRouteBindings("plugin-demo", []pluginhost.SourceRouteBinding{
		{
			PluginID: "plugin-demo",
			Method:   "GET",
			Path:     "/api/v1/plugins/plugin-demo/summary",
		},
	})
	bindings := second.ListSourceRouteBindings()
	if len(bindings) != 1 {
		t.Fatalf("expected second instance to observe 1 shared route binding, got %d", len(bindings))
	}
	if bindings[0].PluginID != "plugin-demo" || bindings[0].Path != "/api/v1/plugins/plugin-demo/summary" {
		t.Fatalf("unexpected shared route binding: %#v", bindings[0])
	}

	second.DeletePluginEnabledState("plugin-demo")
	if firstChecker("plugin-demo") {
		t.Fatal("expected deleting shared snapshot entry to affect all instances")
	}
}
