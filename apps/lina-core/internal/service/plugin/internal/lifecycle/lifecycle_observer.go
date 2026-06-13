// This file stores synchronous host-side plugin lifecycle observers for one
// lifecycle service instance.

package lifecycle

import (
	"context"
	"sort"
	"sync"
)

// lifecycleObserverRegistry stores lifecycle observers for one lifecycle service instance.
type lifecycleObserverRegistry struct {
	mu           sync.RWMutex
	nextID       int
	observerByID map[int]LifecycleObserver
}

// newLifecycleObserverRegistry creates an empty lifecycle observer registry.
func newLifecycleObserverRegistry() *lifecycleObserverRegistry {
	return &lifecycleObserverRegistry{
		observerByID: make(map[int]LifecycleObserver),
	}
}

// RegisterLifecycleObserver subscribes one synchronous lifecycle observer and
// returns its unsubscribe function.
func (s *serviceImpl) RegisterLifecycleObserver(observer LifecycleObserver) func() {
	if s == nil {
		return func() {}
	}
	return s.lifecycleObservers.register(observer)
}

// register subscribes one synchronous lifecycle observer and returns its
// unsubscribe function.
func (r *lifecycleObserverRegistry) register(observer LifecycleObserver) func() {
	if observer == nil {
		return func() {}
	}
	if r == nil {
		return func() {}
	}

	r.mu.Lock()
	r.nextID++
	currentID := r.nextID
	r.observerByID[currentID] = observer
	r.mu.Unlock()

	return func() {
		r.mu.Lock()
		delete(r.observerByID, currentID)
		r.mu.Unlock()
	}
}

// snapshot clones all registered observers for one callback dispatch.
func (r *lifecycleObserverRegistry) snapshot() []LifecycleObserver {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]int, 0, len(r.observerByID))
	for id := range r.observerByID {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	observers := make([]LifecycleObserver, 0, len(ids))
	for _, id := range ids {
		observer := r.observerByID[id]
		if observer == nil {
			continue
		}
		observers = append(observers, observer)
	}
	return observers
}

// notifyPluginInstalled dispatches one successful install transition to all observers.
func (s *serviceImpl) notifyPluginInstalled(ctx context.Context, pluginID string) error {
	for _, observer := range s.lifecycleObservers.snapshot() {
		if err := observer.OnPluginInstalled(ctx, pluginID); err != nil {
			return err
		}
	}
	return nil
}

// notifyPluginEnabled dispatches one successful enable transition to all observers.
func (s *serviceImpl) notifyPluginEnabled(ctx context.Context, pluginID string) error {
	for _, observer := range s.lifecycleObservers.snapshot() {
		if err := observer.OnPluginEnabled(ctx, pluginID); err != nil {
			return err
		}
	}
	return nil
}

// notifyPluginDisabled dispatches one successful disable transition to all observers.
func (s *serviceImpl) notifyPluginDisabled(ctx context.Context, pluginID string) error {
	for _, observer := range s.lifecycleObservers.snapshot() {
		if err := observer.OnPluginDisabled(ctx, pluginID); err != nil {
			return err
		}
	}
	return nil
}

// notifyPluginUninstalled dispatches one successful uninstall transition to all observers.
func (s *serviceImpl) notifyPluginUninstalled(ctx context.Context, pluginID string) error {
	for _, observer := range s.lifecycleObservers.snapshot() {
		if err := observer.OnPluginUninstalled(ctx, pluginID); err != nil {
			return err
		}
	}
	return nil
}
