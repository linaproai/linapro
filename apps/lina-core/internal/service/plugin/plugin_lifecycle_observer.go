// This file stores synchronous host-side plugin lifecycle observers used by
// other host components such as the scheduled-job handler registry.

package plugin

import (
	"context"
	"sort"
	"sync"

	"lina-core/internal/service/plugin/internal/lifecycle"
)

// LifecycleObserver receives synchronous plugin lifecycle callbacks from the host plugin service.
type LifecycleObserver = lifecycle.LifecycleObserver

// LifecycleObserverRegistrar subscribes synchronous lifecycle observers to one
// plugin service instance.
type LifecycleObserverRegistrar interface {
	// RegisterLifecycleObserver subscribes one synchronous lifecycle observer and
	// returns its unsubscribe function.
	RegisterLifecycleObserver(observer LifecycleObserver) func()
}

// lifecycleObserverRegistry stores lifecycle observers for one plugin service
// instance.
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
	lifecycleUnsubscribe := func() {}
	if s.lifecycleSvc != nil {
		lifecycleUnsubscribe = s.lifecycleSvc.RegisterLifecycleObserver(observer)
	}
	if s.lifecycleObservers == nil {
		s.lifecycleObservers = newLifecycleObserverRegistry()
	}
	rootUnsubscribe := s.lifecycleObservers.register(observer)
	return func() {
		lifecycleUnsubscribe()
		rootUnsubscribe()
	}
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

// snapshotLifecycleObservers clones all registered observers for one callback dispatch.
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
	if s == nil || s.lifecycleObservers == nil {
		return nil
	}
	for _, observer := range s.lifecycleObservers.snapshot() {
		if err := observer.OnPluginInstalled(ctx, pluginID); err != nil {
			return err
		}
	}
	return nil
}

// notifyPluginEnabled dispatches one successful enable transition to all observers.
func (s *serviceImpl) notifyPluginEnabled(ctx context.Context, pluginID string) error {
	if s == nil || s.lifecycleObservers == nil {
		return nil
	}
	for _, observer := range s.lifecycleObservers.snapshot() {
		if err := observer.OnPluginEnabled(ctx, pluginID); err != nil {
			return err
		}
	}
	return nil
}

// notifyPluginDisabled dispatches one successful disable transition to all observers.
func (s *serviceImpl) notifyPluginDisabled(ctx context.Context, pluginID string) error {
	if s == nil || s.lifecycleObservers == nil {
		return nil
	}
	for _, observer := range s.lifecycleObservers.snapshot() {
		if err := observer.OnPluginDisabled(ctx, pluginID); err != nil {
			return err
		}
	}
	return nil
}

// notifyPluginUninstalled dispatches one successful uninstall transition to all observers.
func (s *serviceImpl) notifyPluginUninstalled(ctx context.Context, pluginID string) error {
	if s == nil || s.lifecycleObservers == nil {
		return nil
	}
	for _, observer := range s.lifecycleObservers.snapshot() {
		if err := observer.OnPluginUninstalled(ctx, pluginID); err != nil {
			return err
		}
	}
	return nil
}
