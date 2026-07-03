// This file keeps host-service registry introspection helpers scoped to tests.

package wasm

// registeredMethods returns registered service/method pairs in insertion order.
func (r *hostServiceDispatchRegistry) registeredMethods() []hostServiceDispatchMethod {
	if r == nil || len(r.methods) == 0 {
		return nil
	}
	methods := make([]hostServiceDispatchMethod, len(r.methods))
	copy(methods, r.methods)
	return methods
}
