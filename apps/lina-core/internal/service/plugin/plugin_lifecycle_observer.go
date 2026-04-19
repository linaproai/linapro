// This file stores synchronous host-side plugin lifecycle observers used by
// other host components such as the scheduled-job handler registry.

package plugin

import (
	"context"
	"sync"
)

// LifecycleObserver receives synchronous plugin lifecycle callbacks from the
// host plugin service.
type LifecycleObserver interface {
	// OnPluginEnabled handles one successful plugin enable transition.
	OnPluginEnabled(ctx context.Context, pluginID string) error
	// OnPluginDisabled handles one successful plugin disable transition.
	OnPluginDisabled(ctx context.Context, pluginID string) error
	// OnPluginUninstalled handles one successful plugin uninstall transition.
	OnPluginUninstalled(ctx context.Context, pluginID string) error
}

var (
	lifecycleObserverMu   sync.RWMutex
	lifecycleObserverID   int
	lifecycleObserverByID = make(map[int]LifecycleObserver)
)

// RegisterLifecycleObserver subscribes one synchronous lifecycle observer and
// returns its unsubscribe function.
func RegisterLifecycleObserver(observer LifecycleObserver) func() {
	if observer == nil {
		return func() {}
	}

	lifecycleObserverMu.Lock()
	lifecycleObserverID++
	currentID := lifecycleObserverID
	lifecycleObserverByID[currentID] = observer
	lifecycleObserverMu.Unlock()

	return func() {
		lifecycleObserverMu.Lock()
		delete(lifecycleObserverByID, currentID)
		lifecycleObserverMu.Unlock()
	}
}

// snapshotLifecycleObservers clones all registered observers for one callback dispatch.
func snapshotLifecycleObservers() []LifecycleObserver {
	lifecycleObserverMu.RLock()
	defer lifecycleObserverMu.RUnlock()

	observers := make([]LifecycleObserver, 0, len(lifecycleObserverByID))
	for _, observer := range lifecycleObserverByID {
		if observer == nil {
			continue
		}
		observers = append(observers, observer)
	}
	return observers
}

// notifyPluginEnabled dispatches one successful enable transition to all observers.
func notifyPluginEnabled(ctx context.Context, pluginID string) error {
	for _, observer := range snapshotLifecycleObservers() {
		if err := observer.OnPluginEnabled(ctx, pluginID); err != nil {
			return err
		}
	}
	return nil
}

// notifyPluginDisabled dispatches one successful disable transition to all observers.
func notifyPluginDisabled(ctx context.Context, pluginID string) error {
	for _, observer := range snapshotLifecycleObservers() {
		if err := observer.OnPluginDisabled(ctx, pluginID); err != nil {
			return err
		}
	}
	return nil
}

// notifyPluginUninstalled dispatches one successful uninstall transition to all observers.
func notifyPluginUninstalled(ctx context.Context, pluginID string) error {
	for _, observer := range snapshotLifecycleObservers() {
		if err := observer.OnPluginUninstalled(ctx, pluginID); err != nil {
			return err
		}
	}
	return nil
}
