// Package tenantfilter exposes shared tenant query helpers for source plugins
// whose plugin-owned tables use the conventional tenant_id discriminator.
package tenantfilter

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginservice/contract"
)

// serviceImpl implements the tenant filter helper service.
type serviceImpl struct {
	bizCtxSvc       contract.BizCtxService
	bypassEvaluator contract.PlatformBypassEvaluator
}

// New creates tenant filtering helpers from host-owned context dependencies.
func New(
	bizCtxSvc contract.BizCtxService,
	bypassEvaluator contract.PlatformBypassEvaluator,
) (contract.TenantFilterService, error) {
	if bizCtxSvc == nil {
		return nil, gerror.New("tenantfilter requires host bizctx service")
	}
	return &serviceImpl{
		bizCtxSvc:       bizCtxSvc,
		bypassEvaluator: bypassEvaluator,
	}, nil
}
