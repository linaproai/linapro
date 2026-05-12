// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package tenant

import (
	"context"

	"lina-plugin-multi-tenant/backend/api/tenant/v1"
)

type ITenantV1 interface {
	TenantPluginList(ctx context.Context, req *v1.TenantPluginListReq) (res *v1.TenantPluginListRes, err error)
	TenantPluginDisable(ctx context.Context, req *v1.TenantPluginDisableReq) (res *v1.TenantPluginDisableRes, err error)
	TenantPluginEnable(ctx context.Context, req *v1.TenantPluginEnableReq) (res *v1.TenantPluginEnableRes, err error)
}
