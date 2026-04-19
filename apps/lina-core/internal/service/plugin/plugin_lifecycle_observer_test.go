// This file verifies synchronous lifecycle observer subscription and dispatch.

package plugin

import (
	"context"
	"testing"
)

// lifecycleObserverRecorder captures lifecycle events for observer tests.
type lifecycleObserverRecorder struct {
	events []string
}

// OnPluginEnabled records one plugin-enabled event.
func (r *lifecycleObserverRecorder) OnPluginEnabled(ctx context.Context, pluginID string) error {
	r.events = append(r.events, "enabled:"+pluginID)
	return nil
}

// OnPluginDisabled records one plugin-disabled event.
func (r *lifecycleObserverRecorder) OnPluginDisabled(ctx context.Context, pluginID string) error {
	r.events = append(r.events, "disabled:"+pluginID)
	return nil
}

// OnPluginUninstalled records one plugin-uninstalled event.
func (r *lifecycleObserverRecorder) OnPluginUninstalled(ctx context.Context, pluginID string) error {
	r.events = append(r.events, "uninstalled:"+pluginID)
	return nil
}

// TestRegisterLifecycleObserverDispatchesCallbacks verifies lifecycle events
// are delivered synchronously and unsubscribe stops future deliveries.
func TestRegisterLifecycleObserverDispatchesCallbacks(t *testing.T) {
	observer := &lifecycleObserverRecorder{}
	unsubscribe := RegisterLifecycleObserver(observer)

	if err := notifyPluginEnabled(context.Background(), "plugin-demo"); err != nil {
		t.Fatalf("expected enabled notification to succeed, got error: %v", err)
	}
	if err := notifyPluginDisabled(context.Background(), "plugin-demo"); err != nil {
		t.Fatalf("expected disabled notification to succeed, got error: %v", err)
	}
	if err := notifyPluginUninstalled(context.Background(), "plugin-demo"); err != nil {
		t.Fatalf("expected uninstall notification to succeed, got error: %v", err)
	}

	unsubscribe()

	if err := notifyPluginEnabled(context.Background(), "plugin-after-unsubscribe"); err != nil {
		t.Fatalf("expected post-unsubscribe notification to succeed, got error: %v", err)
	}

	expected := []string{
		"enabled:plugin-demo",
		"disabled:plugin-demo",
		"uninstalled:plugin-demo",
	}
	if len(observer.events) != len(expected) {
		t.Fatalf("expected events %#v, got %#v", expected, observer.events)
	}
	for index, item := range expected {
		if observer.events[index] != item {
			t.Fatalf("expected events %#v, got %#v", expected, observer.events)
		}
	}
}
