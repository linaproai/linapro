// This file implements generic plugin resource list routing and
// controller-level permission enforcement for plugin-owned resources.

package plugin

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/api/plugin/v1"
	i18nsvc "lina-core/internal/service/i18n"
	middlewaresvc "lina-core/internal/service/middleware"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/pkg/bizerr"
)

// pluginResourcePermissionErrorResponse mirrors the unified response envelope
// for controller-written plugin resource permission failures.
type pluginResourcePermissionErrorResponse struct {
	Code          int            `json:"code"`
	Message       string         `json:"message"`
	Data          any            `json:"data"`
	ErrorCode     string         `json:"errorCode,omitempty"`
	MessageKey    string         `json:"messageKey,omitempty"`
	MessageParams map[string]any `json:"messageParams,omitempty"`
}

// ResourceList queries plugin-owned backend resources through the generic resource contract.
func (c *ControllerV1) ResourceList(ctx context.Context, req *v1.ResourceListReq) (res *v1.ResourceListRes, err error) {
	allowed, err := c.ensurePluginResourcePermission(ctx, req.Id, req.Resource)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, nil
	}

	filters := make(map[string]string)
	if r := g.RequestFromCtx(ctx); r != nil {
		for key, value := range r.GetQueryMap() {
			if key == "id" || key == "resource" || key == "pageNum" || key == "pageSize" {
				continue
			}
			filters[key] = gconv.String(value)
		}
	}

	out, err := c.pluginSvc.ListResourceRecords(ctx, pluginsvc.ResourceListInput{
		PluginID:   req.Id,
		ResourceID: req.Resource,
		Filters:    filters,
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ResourceListRes{List: out.List, Total: out.Total}, nil
}

// ensurePluginResourcePermission checks the current user's permission set
// against the resolved plugin resource permission string.
func (c *ControllerV1) ensurePluginResourcePermission(
	ctx context.Context,
	pluginID string,
	resourceID string,
) (bool, error) {
	businessCtx := c.bizCtxSvc.Get(ctx)
	if businessCtx == nil || businessCtx.UserId <= 0 {
		writePluginResourcePermissionError(
			ctx,
			c.i18nSvc,
			http.StatusUnauthorized,
			bizerr.NewCode(middlewaresvc.CodeMiddlewarePermissionCurrentUserMissing),
		)
		return false, nil
	}

	accessContext, err := c.roleSvc.GetUserAccessContext(ctx, businessCtx.UserId)
	if err != nil {
		return false, bizerr.WrapCode(err, middlewaresvc.CodeMiddlewarePermissionContextLoadFailed)
	}
	if accessContext != nil {
		c.bizCtxSvc.SetUserAccess(
			ctx,
			int(accessContext.DataScope),
			accessContext.DataScopeUnsupported,
			accessContext.UnsupportedDataScope,
		)
	}

	requiredPermission, err := c.pluginSvc.ResolveResourcePermission(ctx, pluginID, resourceID)
	if err != nil {
		return false, err
	}
	if hasPluginResourcePermission(accessContext, requiredPermission) {
		return true, nil
	}

	writePluginResourcePermissionError(
		ctx,
		c.i18nSvc,
		http.StatusForbidden,
		bizerr.NewCode(
			middlewaresvc.CodeMiddlewarePermissionDeniedRequired,
			bizerr.P("permissions", requiredPermission),
		),
	)
	return false, nil
}

// hasPluginResourcePermission reports whether the resolved permission is empty,
// granted explicitly, or covered by a super-admin/wildcard grant.
func hasPluginResourcePermission(accessContext *role.UserAccessContext, requiredPermission string) bool {
	if strings.TrimSpace(requiredPermission) == "" {
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
	if _, ok := granted["*:*:*"]; ok {
		return true
	}
	_, ok := granted[strings.TrimSpace(requiredPermission)]
	return ok
}

// writePluginResourcePermissionError writes one standardized JSON permission
// error response onto the current request context and aborts processing.
func writePluginResourcePermissionError(
	ctx context.Context,
	i18nSvc i18nsvc.Translator,
	status int,
	err error,
) {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		return
	}

	message := ""
	if i18nSvc != nil {
		message = i18nSvc.LocalizeError(ctx, err)
	}
	if message == "" && err != nil {
		message = err.Error()
	}

	r.SetError(err)
	r.Response.WriteStatus(status)
	response := pluginResourcePermissionErrorResponse{
		Code:    gcode.CodeUnknown.Code(),
		Data:    nil,
		Message: message,
	}
	applyPluginResourcePermissionErrorMetadata(&response, err)
	r.Response.WriteJson(response)
	r.ExitAll()
}

// applyPluginResourcePermissionErrorMetadata copies structured runtime-message
// metadata into the controller-written plugin resource permission response.
func applyPluginResourcePermissionErrorMetadata(response *pluginResourcePermissionErrorResponse, err error) {
	if response == nil || err == nil {
		return
	}
	messageErr, ok := bizerr.As(err)
	if !ok {
		return
	}
	response.Code = messageErr.TypeCode().Code()
	response.ErrorCode = messageErr.RuntimeCode()
	response.MessageKey = messageErr.MessageKey()
	response.MessageParams = messageErr.Params()
}
