// Package middleware implements monitor-operlog HTTP audit middleware and
// request normalization services for the source plugin.
package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"

	hostaudit "lina-core/pkg/pluginservice/audit"
	hostbizctx "lina-core/pkg/pluginservice/bizctx"
)

// Service defines the monitor-operlog middleware service contract.
type Service interface {
	// Audit captures one completed request and dispatches the normalized audit event.
	Audit(request *ghttp.Request)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	auditSvc  hostaudit.Service  // audit event publisher and dynamic-route metadata reader
	bizCtxSvc hostbizctx.Service // authenticated operator identity reader
}

// New creates and returns a new monitor-operlog middleware service instance.
func New() Service {
	return newWithServices(hostaudit.New(), hostbizctx.New())
}

// newWithServices creates one middleware service with explicit host service dependencies.
func newWithServices(auditSvc hostaudit.Service, bizCtxSvc hostbizctx.Service) Service {
	return &serviceImpl{
		auditSvc:  auditSvc,
		bizCtxSvc: bizCtxSvc,
	}
}
