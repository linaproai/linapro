// This file handles public password-recovery email requests.

package auth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/auth/v1"
	authsvc "lina-core/internal/service/auth"
)

// ForgetPassword handles public password-recovery requests.
func (c *ControllerV1) ForgetPassword(ctx context.Context, req *v1.ForgetPasswordReq) (res *v1.ForgetPasswordRes, err error) {
	if err = c.authSvc.RequestPasswordReset(ctx, authsvc.PasswordResetRequestInput{
		Email:             req.Email,
		PublicOrigin:      resolvePublicOrigin(ctx),
		WorkspaceBasePath: c.configSvc.GetWorkspaceBasePath(ctx),
	}); err != nil {
		return nil, err
	}
	return &v1.ForgetPasswordRes{Accepted: true}, nil
}

// resolvePublicOrigin builds the browser-facing origin for recovery links.
func resolvePublicOrigin(ctx context.Context) string {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		return ""
	}
	scheme := "http"
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		scheme = strings.ToLower(strings.Split(forwarded, ",")[0])
	} else if r.TLS != nil {
		scheme = "https"
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return ""
	}
	return scheme + "://" + host
}
