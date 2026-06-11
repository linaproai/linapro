// This file exposes shared tenant query helpers for source plugins
// whose plugin-owned tables use the conventional tenant_id discriminator.
package tenantspi

import (
	"lina-core/pkg/plugin/capability/bizctxcap"

	"github.com/gogf/gf/v2/errors/gerror"
)

// pluginTableFilterService implements the tenant filter helper service.
type pluginTableFilterService struct {
	bizCtxSvc       bizctxcap.Service
	bypassEvaluator PlatformBypassEvaluator
}

// NewPluginTableFilter creates tenant filtering helpers from host-owned context dependencies.
func NewPluginTableFilter(
	bizCtxSvc bizctxcap.Service,
	bypassEvaluator PlatformBypassEvaluator,
) (PluginTableFilterService, error) {
	if bizCtxSvc == nil {
		return nil, gerror.New("tenantfilter requires host bizctx service")
	}
	return &pluginTableFilterService{
		bizCtxSvc:       bizCtxSvc,
		bypassEvaluator: bypassEvaluator,
	}, nil
}
