// This file exposes the install-mock-data opt-in flag through context so
// downstream lifecycle and runtime packages can detect the operator's choice
// without adding another parameter to every reconciler/install method.

package catalog

import "context"

// installMockDataContextKey is the private context-key type for the
// install-mock-data opt-in flag. Using a typed key prevents collisions with
// unrelated context values throughout the install/reconcile call chain.
type installMockDataContextKey struct{}

// WithInstallMockData decorates ctx with the operator's install-mock-data
// opt-in choice. Callers in the plugin facade set this once at the entry
// point so downstream code (source plugin install, dynamic plugin
// reconciler) can read it without threading a bool argument.
func WithInstallMockData(ctx context.Context, enable bool) context.Context {
	return context.WithValue(ctx, installMockDataContextKey{}, enable)
}

// ShouldInstallMockData reports whether the current request was decorated
// with an install-mock-data opt-in. The default is false, preserving the
// pre-existing install behavior for any caller that did not opt in.
func ShouldInstallMockData(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	v, ok := ctx.Value(installMockDataContextKey{}).(bool)
	return ok && v
}
