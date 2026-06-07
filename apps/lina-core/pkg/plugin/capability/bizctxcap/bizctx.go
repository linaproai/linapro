// This file defines the business-context service implementation used by source
// plugins so they can read current request identity, tenancy, and impersonation
// metadata without depending on host-internal service packages.
package bizctxcap

// serviceAdapter reads plugin-visible context from an optional provider or context value.
type serviceAdapter struct {
	provider ContextProvider
}

// New creates and returns a business-context service backed by the optional provider.
func New(provider ContextProvider) Service {
	return &serviceAdapter{provider: provider}
}
