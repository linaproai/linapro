// This file implements declarative permission enforcement for static host APIs.

package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/role"
)

const (
	staticPermissionMetaTag      = "permission"
	staticPermissionMetaTagAlias = "perms"
	staticPermissionWildcard     = "*:*:*"
	staticPermissionErrorCode    = 1
)

type permissionErrorResponse struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

// Permission enforces declarative permission requirements declared on static host API handlers.
func (s *serviceImpl) Permission(r *ghttp.Request) {
	if r == nil {
		return
	}

	requiredPermissions := extractDeclaredPermissions(r)
	if len(requiredPermissions) == 0 {
		// Build-time audit tests ensure protected static APIs declare permissions.
		// Middleware therefore treats "no metadata" as "no extra permission gate".
		r.Middleware.Next()
		return
	}

	businessCtx := s.bizCtxSvc.Get(r.Context())
	if businessCtx == nil || businessCtx.UserId <= 0 {
		writePermissionError(
			r,
			http.StatusUnauthorized,
			gerror.New("未获取到当前登录用户"),
			"未获取到当前登录用户",
		)
		return
	}

	accessContext, err := s.roleSvc.GetUserAccessContext(r.Context(), businessCtx.UserId)
	if err != nil {
		writePermissionError(
			r,
			http.StatusInternalServerError,
			gerror.Wrap(err, "加载接口权限上下文失败"),
			"加载接口权限上下文失败",
		)
		return
	}
	if hasRequiredPermissions(accessContext, requiredPermissions) {
		r.Middleware.Next()
		return
	}

	message := permissionDeniedMessage(requiredPermissions)
	writePermissionError(r, http.StatusForbidden, gerror.New(message), message)
}

// extractDeclaredPermissions reads the permission metadata declared on the
// current request DTO/handler and normalizes it into one deduplicated list.
func extractDeclaredPermissions(r *ghttp.Request) []string {
	if r == nil {
		return nil
	}
	handler := r.GetServeHandler()
	if handler != nil {
		permissions := resolveDeclaredPermissions(
			handler.GetMetaTag(staticPermissionMetaTag),
			handler.GetMetaTag(staticPermissionMetaTagAlias),
		)
		if len(permissions) > 0 {
			return permissions
		}
	}
	return nil
}

// resolveDeclaredPermissions prefers the canonical permission tag and falls
// back to the legacy alias so older declarations remain compatible.
func resolveDeclaredPermissions(permissionTag string, aliasTag string) []string {
	permissions := normalizePermissionList(permissionTag)
	if len(permissions) > 0 {
		return permissions
	}
	return normalizePermissionList(aliasTag)
}

// normalizePermissionList trims, deduplicates, and preserves order for the
// comma-separated permission list declared in route metadata.
func normalizePermissionList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var (
		parts  = strings.Split(raw, ",")
		result = make([]string, 0, len(parts))
		seen   = make(map[string]struct{}, len(parts))
	)
	for _, part := range parts {
		permission := strings.TrimSpace(part)
		if permission == "" {
			continue
		}
		if _, ok := seen[permission]; ok {
			continue
		}
		seen[permission] = struct{}{}
		result = append(result, permission)
	}
	return result
}

// hasRequiredPermissions applies the static-host permission semantics: super
// admin and wildcard bypass, otherwise every declared permission must be granted.
func hasRequiredPermissions(accessContext *role.UserAccessContext, required []string) bool {
	if len(required) == 0 {
		return true
	}
	if accessContext == nil {
		return false
	}
	if accessContext.IsSuperAdmin {
		return true
	}

	granted := make(map[string]struct{}, len(accessContext.Permissions))
	for _, permission := range accessContext.Permissions {
		currentPermission := strings.TrimSpace(permission)
		if currentPermission == "" {
			continue
		}
		granted[currentPermission] = struct{}{}
	}
	if _, ok := granted[staticPermissionWildcard]; ok {
		return true
	}

	for _, permission := range required {
		if _, ok := granted[permission]; !ok {
			return false
		}
	}
	return true
}

// permissionDeniedMessage formats the operator-facing permission denial message.
func permissionDeniedMessage(required []string) string {
	return "当前用户缺少接口权限: " + strings.Join(required, ", ")
}

// writePermissionError writes one JSON error payload and binds the error onto
// the request so upper layers can still observe the failure cause.
func writePermissionError(r *ghttp.Request, status int, err error, message string) {
	r.SetError(err)
	r.Response.WriteStatus(status)
	r.Response.WriteJson(permissionErrorResponse{
		Code:    staticPermissionErrorCode,
		Data:    nil,
		Message: message,
	})
}
