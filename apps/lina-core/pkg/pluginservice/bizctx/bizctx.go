// Package bizctx exposes a narrowed view of the host business context to source
// plugins so they can read current request identity, tenancy, and impersonation
// metadata without depending on host-internal service packages.
package bizctx

import "lina-core/pkg/pluginservice/contract"

// serviceAdapter reads plugin-visible context from an optional provider or context value.
type serviceAdapter struct {
	provider contract.ContextProvider
}

// New creates and returns a business-context service backed by the optional provider.
func New(provider contract.ContextProvider) contract.BizCtxService {
	return &serviceAdapter{provider: provider}
}
