// This file implements tenant-resolution middleware for protected requests.

package middleware

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// Tenancy resolves tenant identity and injects it into request business context.
func (s *serviceImpl) Tenancy(r *ghttp.Request) {
	if r == nil {
		return
	}
	if s == nil || s.tenantSvc == nil || !s.tenantSvc.Enabled(r.Context()) {
		s.bizCtxSvc.SetTenant(r.Context(), int(pkgtenantcap.PLATFORM))
		r.Middleware.Next()
		return
	}

	result, err := s.tenantSvc.ResolveTenant(r.Context(), r)
	if err != nil {
		r.SetError(err)
		status := http.StatusForbidden
		if bizerr.Is(err, pkgtenantcap.CodeTenantRequired) {
			status = http.StatusUnauthorized
		}
		r.Response.WriteStatus(status)
		return
	}
	if result == nil || !result.Matched {
		err = bizerr.NewCode(pkgtenantcap.CodeTenantRequired)
		r.SetError(err)
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	s.bizCtxSvc.SetTenant(r.Context(), int(result.TenantID))
	if result.IsImpersonation || result.ActingAsTenant {
		s.bizCtxSvc.SetImpersonation(
			r.Context(),
			result.ActingUserID,
			int(result.TenantID),
			result.ActingAsTenant,
			result.IsImpersonation,
		)
	}
	r.Middleware.Next()
}
