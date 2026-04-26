// Package middleware implements monitor-operlog HTTP audit middleware and
// request normalization services for the source plugin.
package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"

	hostbizctx "lina-core/pkg/pluginservice/bizctx"
	hostroute "lina-core/pkg/pluginservice/route"
	operlogsvc "lina-plugin-monitor-operlog/backend/internal/service/operlog"
)

// Service defines the monitor-operlog middleware service contract.
type Service interface {
	// Audit captures one completed request and persists the normalized operation log.
	Audit(request *ghttp.Request)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	routeMetaSvc hostroute.Service  // dynamic-route metadata reader
	bizCtxSvc    hostbizctx.Service // authenticated operator identity reader
	operLogSvc   operlogsvc.Service // plugin-owned operation-log persistence service
}

// New creates and returns a new monitor-operlog middleware service instance.
func New() Service {
	return newWithServices(hostroute.New(), hostbizctx.New(), operlogsvc.New())
}

// newWithServices creates one middleware service with explicit host service dependencies.
func newWithServices(
	routeMetaSvc hostroute.Service,
	bizCtxSvc hostbizctx.Service,
	operLogSvc operlogsvc.Service,
) Service {
	return &serviceImpl{
		routeMetaSvc: routeMetaSvc,
		bizCtxSvc:    bizCtxSvc,
		operLogSvc:   operLogSvc,
	}
}
