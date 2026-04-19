package plugin

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
)

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
			http.StatusUnauthorized,
			gerror.New("未获取到当前登录用户"),
			"未获取到当前登录用户",
		)
		return false, nil
	}

	accessContext, err := c.roleSvc.GetUserAccessContext(ctx, businessCtx.UserId)
	if err != nil {
		return false, gerror.Wrap(err, "加载插件资源权限上下文失败")
	}
	requiredPermission, err := c.pluginSvc.ResolveResourcePermission(ctx, pluginID, resourceID)
	if err != nil {
		return false, err
	}
	if hasPluginResourcePermission(accessContext, requiredPermission) {
		return true, nil
	}

	message := "当前用户缺少接口权限: " + requiredPermission
	writePluginResourcePermissionError(ctx, http.StatusForbidden, gerror.New(message), message)
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
	status int,
	err error,
	message string,
) {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		return
	}

	r.SetError(err)
	r.Response.WriteStatus(status)
	r.Response.WriteJson(g.Map{
		"code":    1,
		"data":    nil,
		"message": message,
	})
	r.ExitAll()
}
