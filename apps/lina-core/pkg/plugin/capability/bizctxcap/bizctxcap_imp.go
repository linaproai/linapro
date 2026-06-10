// This file defines the business-context service implementation used by source
// plugins so they can read current request identity, tenancy, and impersonation
// metadata without depending on host-internal service packages.
package bizctxcap

import "context"

// serviceAdapter reads plugin-visible context from an optional provider or context value.
type serviceAdapter struct {
	provider ContextProvider
}

// New creates and returns a business-context service backed by the optional provider.
func New(provider ContextProvider) Service {
	return &serviceAdapter{provider: provider}
}

// Current returns a read-only snapshot of the request context fields published
// to source plugins.
func (s *serviceAdapter) Current(ctx context.Context) CurrentContext {
	if s != nil && s.provider != nil {
		return s.provider.Current(ctx)
	}
	return CurrentFromContext(ctx)
}
