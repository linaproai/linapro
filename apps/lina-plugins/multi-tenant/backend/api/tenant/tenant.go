// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package tenant

import (
	"context"

	"lina-plugin-multi-tenant/backend/api/tenant/v1"
)

type ITenantV1 interface {
	MemberList(ctx context.Context, req *v1.MemberListReq) (res *v1.MemberListRes, err error)
	MemberAdd(ctx context.Context, req *v1.MemberAddReq) (res *v1.MemberAddRes, err error)
	MemberMe(ctx context.Context, req *v1.MemberMeReq) (res *v1.MemberMeRes, err error)
	MemberRemove(ctx context.Context, req *v1.MemberRemoveReq) (res *v1.MemberRemoveRes, err error)
	MemberUpdate(ctx context.Context, req *v1.MemberUpdateReq) (res *v1.MemberUpdateRes, err error)
	TenantPluginList(ctx context.Context, req *v1.TenantPluginListReq) (res *v1.TenantPluginListRes, err error)
	TenantPluginDisable(ctx context.Context, req *v1.TenantPluginDisableReq) (res *v1.TenantPluginDisableRes, err error)
	TenantPluginEnable(ctx context.Context, req *v1.TenantPluginEnableReq) (res *v1.TenantPluginEnableRes, err error)
}
