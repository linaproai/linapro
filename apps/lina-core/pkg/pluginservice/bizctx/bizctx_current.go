// This file contains the concrete business-context lookup behavior for source
// plugins. It keeps provider fallback logic outside the package entrypoint so
// the main file remains a stable construction and contract boundary.

package bizctx

import (
	"context"

	"lina-core/pkg/pluginservice/contract"
)

// Current returns a read-only snapshot of the request context fields published
// to source plugins.
func (s *serviceAdapter) Current(ctx context.Context) contract.CurrentContext {
	if s != nil && s.provider != nil {
		return s.provider.Current(ctx)
	}
	return contract.CurrentFromContext(ctx)
}
