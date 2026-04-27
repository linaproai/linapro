// This file implements declarative permission enforcement for static host APIs.

package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/role"
)

// Permission middleware constants define metadata tag names, wildcard grants,
// and the normalized JSON error envelope code.
const (
	staticPermissionMetaTag      = "permission"
	staticPermissionMetaTagAlias = "perms"
	staticPermissionWildcard     = "*:*:*"
	staticPermissionErrorCode    = 1
)

// permissionErrorResponse defines the JSON payload returned when permission
// middleware rejects a request.
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
			s.i18nSvc,
			http.StatusUnauthorized,
			gerror.New("error.permission.currentUserMissing"),
		)
		return
	}

	accessContext, err := s.roleSvc.GetUserAccessContext(r.Context(), businessCtx.UserId)
	if err != nil {
		writePermissionError(
			r,
			s.i18nSvc,
			http.StatusInternalServerError,
			gerror.Wrap(err, "error.permission.contextLoadFailed"),
		)
		return
	}
	if hasRequiredPermissions(accessContext, requiredPermissions) {
		r.Middleware.Next()
		return
	}

	writePermissionError(
		r,
		s.i18nSvc,
		http.StatusForbidden,
		gerror.Newf("error.permission.denied.required", strings.Join(requiredPermissions, ", ")),
	)
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

// writePermissionError writes one JSON error payload and binds the error onto
// the request so upper layers can still observe the failure cause.
func writePermissionError(r *ghttp.Request, i18nSvc middlewareI18nService, status int, err error) {
	if r == nil {
		return
	}

	message := ""
	if i18nSvc != nil {
		message = i18nSvc.LocalizeError(r.Context(), err)
	}
	if message == "" && err != nil {
		message = err.Error()
	}

	r.SetError(err)
	r.Response.WriteStatus(status)
	r.Response.WriteJson(permissionErrorResponse{
		Code:    staticPermissionErrorCode,
		Data:    nil,
		Message: message,
	})
}
