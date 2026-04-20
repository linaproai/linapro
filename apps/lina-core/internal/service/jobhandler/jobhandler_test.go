// This file verifies handler registry mutation and parameter validation behavior.

package jobhandler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"lina-core/internal/service/jobmeta"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/pluginhost"
)

// testLogCleaner satisfies RegisterHostHandlers for registry tests.
type testLogCleaner struct{}

// CleanupDueLogs is a no-op for host-handler registry tests.
func (testLogCleaner) CleanupDueLogs(ctx context.Context) (int64, error) { return 0, nil }

// TestRegisterRejectsDuplicateRefs verifies handler refs remain globally unique.
func TestRegisterRejectsDuplicateRefs(t *testing.T) {
	registry := New()
	definition := HandlerDef{
		Ref:          "host:test",
		DisplayName:  "Test Handler",
		ParamsSchema: `{"type":"object","properties":{}}`,
		Source:       jobmeta.HandlerSourceHost,
		Invoke: func(ctx context.Context, params json.RawMessage) (result any, err error) {
			return nil, nil
		},
	}

	if err := registry.Register(definition); err != nil {
		t.Fatalf("expected first register call to succeed, got error: %v", err)
	}
	if err := registry.Register(definition); err == nil {
		t.Fatal("expected duplicate register call to fail")
	}
}

// TestLookupAndUnregisterNotify verifies registry lookups and change callbacks stay in sync.
func TestLookupAndUnregisterNotify(t *testing.T) {
	registry := New()
	var notifications []string
	unsubscribe := registry.SubscribeChanges(func(ref string, exists bool) {
		state := "removed"
		if exists {
			state = "registered"
		}
		notifications = append(notifications, ref+":"+state)
	})
	defer unsubscribe()

	definition := HandlerDef{
		Ref:          "host:test-lookup",
		DisplayName:  "Lookup Handler",
		ParamsSchema: `{"type":"object","properties":{}}`,
		Source:       jobmeta.HandlerSourceHost,
		Invoke: func(ctx context.Context, params json.RawMessage) (result any, err error) {
			return map[string]any{"ok": true}, nil
		},
	}
	if err := registry.Register(definition); err != nil {
		t.Fatalf("expected register to succeed, got error: %v", err)
	}

	lookup, ok := registry.Lookup(definition.Ref)
	if !ok || lookup.Ref != definition.Ref {
		t.Fatalf("expected handler lookup to succeed, got ok=%t def=%#v", ok, lookup)
	}

	registry.Unregister(definition.Ref)
	if _, ok = registry.Lookup(definition.Ref); ok {
		t.Fatal("expected handler lookup to miss after unregister")
	}
	if len(notifications) != 2 {
		t.Fatalf("expected register and unregister notifications, got %#v", notifications)
	}
}

// TestValidateParams verifies the supported JSON Schema subset enforces required fields and types.
func TestValidateParams(t *testing.T) {
	schema := `{
		"type":"object",
		"properties":{
			"name":{"type":"string"},
			"count":{"type":"integer"},
			"enabled":{"type":"boolean"}
		},
		"required":["name","count"]
	}`

	if err := ValidateParams(schema, json.RawMessage(`{"name":"demo","count":2,"enabled":true}`)); err != nil {
		t.Fatalf("expected valid params to pass, got error: %v", err)
	}
	if err := ValidateParams(schema, json.RawMessage(`{"name":"demo"}`)); err == nil {
		t.Fatal("expected missing required field to fail validation")
	}
	if err := ValidateParams(schema, json.RawMessage(`{"name":"demo","count":"bad"}`)); err == nil {
		t.Fatal("expected type mismatch to fail validation")
	}
}

// TestRegisterHostHandlersProvidesWaitHandler verifies the built-in wait
// handler is registered and respects execution-context cancellation.
func TestRegisterHostHandlersProvidesWaitHandler(t *testing.T) {
	registry := New()
	if err := RegisterHostHandlers(registry, testLogCleaner{}); err != nil {
		t.Fatalf("expected host handler registration to succeed, got error: %v", err)
	}

	definition, ok := registry.Lookup("host:wait")
	if !ok {
		t.Fatal("expected host:wait handler to be registered")
	}
	if err := ValidateParams(definition.ParamsSchema, json.RawMessage(`{"seconds":1}`)); err != nil {
		t.Fatalf("expected host:wait schema validation to pass, got error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := definition.Invoke(ctx, json.RawMessage(`{"seconds":1}`))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected host:wait handler to honor context cancellation, got error: %v", err)
	}
}

// testPluginStatusChecker exposes enabled-state snapshots for plugin lifecycle tests.
type testPluginStatusChecker struct {
	enabled     map[string]bool
	managedJobs map[string][]pluginsvc.ManagedCronJob
}

// IsEnabled reports whether one plugin is flagged enabled in the test snapshot.
func (c testPluginStatusChecker) IsEnabled(ctx context.Context, pluginID string) bool {
	return c.enabled[pluginID]
}

// ListEnabledPluginIDs returns the enabled plugin IDs for startup lifecycle tests.
func (c testPluginStatusChecker) ListEnabledPluginIDs(ctx context.Context) ([]string, error) {
	items := make([]string, 0, len(c.enabled))
	for pluginID, enabled := range c.enabled {
		if !enabled {
			continue
		}
		items = append(items, pluginID)
	}
	return items, nil
}

// ListManagedCronJobsByPlugin returns no synthetic cron jobs for registry tests
// unless one test overrides the fixture explicitly.
func (c testPluginStatusChecker) ListManagedCronJobsByPlugin(
	ctx context.Context,
	pluginID string,
) ([]pluginsvc.ManagedCronJob, error) {
	return c.managedJobs[pluginID], nil
}

// TestAttachPluginLifecycleSyncsEnabledSourcePluginHandlers verifies startup
// sync registers source-plugin handlers for already-enabled plugins.
func TestAttachPluginLifecycleSyncsEnabledSourcePluginHandlers(t *testing.T) {
	const pluginID = "jobhandler-lifecycle-enabled-sync"

	sourcePlugin := pluginhost.NewSourcePlugin(pluginID)
	sourcePlugin.RegisterJobHandler(pluginhost.JobHandlerRegistration{
		Name:         "echo",
		DisplayName:  "Echo",
		Description:  "Echo handler for lifecycle sync tests.",
		ParamsSchema: `{"type":"object","properties":{"message":{"type":"string"}},"required":["message"]}`,
		Handler: func(ctx context.Context, params json.RawMessage) (result any, err error) {
			return map[string]any{"params": string(params)}, nil
		},
	})
	pluginhost.RegisterSourcePlugin(sourcePlugin)

	registry := New()
	unsubscribe, err := AttachPluginLifecycle(
		context.Background(),
		registry,
		testPluginStatusChecker{enabled: map[string]bool{pluginID: true}},
	)
	if err != nil {
		t.Fatalf("expected plugin lifecycle attachment to succeed, got error: %v", err)
	}
	defer unsubscribe()

	definition, ok := registry.Lookup("plugin:" + pluginID + "/echo")
	if !ok {
		t.Fatal("expected enabled source-plugin handler to be registered during startup sync")
	}
	if definition.Source != jobmeta.HandlerSourcePlugin {
		t.Fatalf("expected plugin handler source, got %s", definition.Source)
	}
	if definition.PluginID != pluginID {
		t.Fatalf("expected plugin id %s, got %s", pluginID, definition.PluginID)
	}
}

// TestPluginLifecycleObserverRegistersAndUnregistersHandlers verifies plugin
// lifecycle callbacks keep registry state in sync with plugin handler ownership.
func TestPluginLifecycleObserverRegistersAndUnregistersHandlers(t *testing.T) {
	const pluginID = "jobhandler-lifecycle-transition"

	sourcePlugin := pluginhost.NewSourcePlugin(pluginID)
	sourcePlugin.RegisterJobHandler(pluginhost.JobHandlerRegistration{
		Name:         "echo",
		DisplayName:  "Echo",
		Description:  "Echo handler for lifecycle transition tests.",
		ParamsSchema: `{"type":"object","properties":{"message":{"type":"string"}},"required":["message"]}`,
		Handler: func(ctx context.Context, params json.RawMessage) (result any, err error) {
			return map[string]any{"params": string(params)}, nil
		},
	})
	pluginhost.RegisterSourcePlugin(sourcePlugin)

	observer := &pluginLifecycleObserver{registry: New()}
	if err := observer.OnPluginEnabled(context.Background(), pluginID); err != nil {
		t.Fatalf("expected plugin enable callback to succeed, got error: %v", err)
	}
	if _, ok := observer.registry.Lookup("plugin:" + pluginID + "/echo"); !ok {
		t.Fatal("expected plugin handler to be registered after enable callback")
	}

	if err := observer.OnPluginDisabled(context.Background(), pluginID); err != nil {
		t.Fatalf("expected plugin disable callback to succeed, got error: %v", err)
	}
	if _, ok := observer.registry.Lookup("plugin:" + pluginID + "/echo"); ok {
		t.Fatal("expected plugin handler to be removed after disable callback")
	}

	if err := observer.OnPluginEnabled(context.Background(), pluginID); err != nil {
		t.Fatalf("expected plugin re-enable callback to succeed, got error: %v", err)
	}
	if _, ok := observer.registry.Lookup("plugin:" + pluginID + "/echo"); !ok {
		t.Fatal("expected plugin handler to be re-registered after re-enable callback")
	}

	if err := observer.OnPluginUninstalled(context.Background(), pluginID); err != nil {
		t.Fatalf("expected plugin uninstall callback to succeed, got error: %v", err)
	}
	if _, ok := observer.registry.Lookup("plugin:" + pluginID + "/echo"); ok {
		t.Fatal("expected plugin handler to be removed after uninstall callback")
	}
}

// TestAttachPluginLifecycleSyncsEnabledDynamicPluginCronHandlers verifies
// startup sync also restores synthetic handlers for enabled dynamic plugins.
func TestAttachPluginLifecycleSyncsEnabledDynamicPluginCronHandlers(t *testing.T) {
	const pluginID = "jobhandler-dynamic-enabled-sync"

	registry := New()
	unsubscribe, err := AttachPluginLifecycle(
		context.Background(),
		registry,
		testPluginStatusChecker{
			enabled: map[string]bool{pluginID: true},
			managedJobs: map[string][]pluginsvc.ManagedCronJob{
				pluginID: {
					{
						PluginID:    pluginID,
						Name:        "heartbeat",
						DisplayName: "Heartbeat",
						Description: "Dynamic cron heartbeat handler.",
						Handler: func(ctx context.Context) error {
							return nil
						},
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("expected plugin lifecycle attachment to succeed, got error: %v", err)
	}
	defer unsubscribe()

	definition, ok := registry.Lookup("plugin:" + pluginID + "/cron:heartbeat")
	if !ok {
		t.Fatal("expected enabled dynamic-plugin cron handler to be registered during startup sync")
	}
	if definition.Source != jobmeta.HandlerSourcePlugin {
		t.Fatalf("expected plugin handler source, got %s", definition.Source)
	}
	if definition.PluginID != pluginID {
		t.Fatalf("expected plugin id %s, got %s", pluginID, definition.PluginID)
	}
}
