// This file defines host plugin lifecycle orchestration through the stable
// source-plugin service contract.
package plugincap

// service delegates lifecycle orchestration to the host-owned runner.
type service struct {
	runner LifecycleRunner
}

// NewLifecycle creates a source-plugin visible plugin lifecycle service.
func NewLifecycle(runner LifecycleRunner) LifecycleService {
	return &service{runner: runner}
}
