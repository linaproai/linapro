// Package middleware implements the demo-control request-guard middleware.
package middleware

import "github.com/gogf/gf/v2/net/ghttp"

// Service defines the demo-control middleware service contract.
type Service interface {
	// Guard enforces the demo-mode read-only policy on API requests.
	Guard(request *ghttp.Request)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

// New creates and returns a new demo-control middleware service.
func New() Service {
	return &serviceImpl{}
}
