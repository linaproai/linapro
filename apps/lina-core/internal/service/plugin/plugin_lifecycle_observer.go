// This file exposes lifecycle observer subscription through the root plugin facade.

package plugin

import "lina-core/internal/service/plugin/internal/lifecycle"

// LifecycleObserver receives synchronous plugin lifecycle callbacks from the host plugin service.
type LifecycleObserver = lifecycle.LifecycleObserver

// RegisterLifecycleObserver subscribes one synchronous lifecycle observer and
// returns its unsubscribe function.
func (s *serviceImpl) RegisterLifecycleObserver(observer LifecycleObserver) func() {
	if s == nil || s.lifecycleSvc == nil {
		return func() {}
	}
	return s.lifecycleSvc.RegisterLifecycleObserver(observer)
}
