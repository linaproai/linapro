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

// OnPluginInstalled records one plugin-installed event.
func (r *lifecycleObserverRecorder) OnPluginInstalled(ctx context.Context, pluginID string) error {
	r.events = append(r.events, "installed:"+pluginID)
	return nil
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

	if err := notifyPluginInstalled(context.Background(), "plugin-demo"); err != nil {
		t.Fatalf("expected install notification to succeed, got error: %v", err)
	}
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
		"installed:plugin-demo",
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

// TestLifecycleObserverDispatchPreservesRegistrationOrder verifies later
// observers receive callbacks after earlier registrations.
func TestLifecycleObserverDispatchPreservesRegistrationOrder(t *testing.T) {
	events := make([]string, 0, 2)

	first := &orderedLifecycleObserverRecorder{
		name:   "first",
		events: &events,
	}
	second := &orderedLifecycleObserverRecorder{
		name:   "second",
		events: &events,
	}
	unsubscribeFirst := RegisterLifecycleObserver(first)
	defer unsubscribeFirst()
	unsubscribeSecond := RegisterLifecycleObserver(second)
	defer unsubscribeSecond()

	if err := notifyPluginEnabled(context.Background(), "plugin-ordered"); err != nil {
		t.Fatalf("expected ordered notification to succeed, got error: %v", err)
	}

	expected := []string{
		"first:enabled:plugin-ordered",
		"second:enabled:plugin-ordered",
	}
	if len(events) != len(expected) {
		t.Fatalf("expected ordered events %#v, got %#v", expected, events)
	}
	for index, item := range expected {
		if events[index] != item {
			t.Fatalf("expected ordered events %#v, got %#v", expected, events)
		}
	}
}

// orderedLifecycleObserverRecorder appends callback order into the shared event slice.
type orderedLifecycleObserverRecorder struct {
	name   string
	events *[]string
}

// OnPluginInstalled records one ordered install event.
func (r *orderedLifecycleObserverRecorder) OnPluginInstalled(ctx context.Context, pluginID string) error {
	*r.events = append(*r.events, r.name+":installed:"+pluginID)
	return nil
}

// OnPluginEnabled records one ordered enable event.
func (r *orderedLifecycleObserverRecorder) OnPluginEnabled(ctx context.Context, pluginID string) error {
	*r.events = append(*r.events, r.name+":enabled:"+pluginID)
	return nil
}

// OnPluginDisabled records one ordered disable event.
func (r *orderedLifecycleObserverRecorder) OnPluginDisabled(ctx context.Context, pluginID string) error {
	*r.events = append(*r.events, r.name+":disabled:"+pluginID)
	return nil
}

// OnPluginUninstalled records one ordered uninstall event.
func (r *orderedLifecycleObserverRecorder) OnPluginUninstalled(ctx context.Context, pluginID string) error {
	*r.events = append(*r.events, r.name+":uninstalled:"+pluginID)
	return nil
}
