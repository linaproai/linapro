// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package auth

import (
	"context"

	"lina-plugin-multi-tenant/backend/api/auth/v1"
)

type IAuthV1 interface {
	LoginTenants(ctx context.Context, req *v1.LoginTenantsReq) (res *v1.LoginTenantsRes, err error)
	SelectTenant(ctx context.Context, req *v1.SelectTenantReq) (res *v1.SelectTenantRes, err error)
	SwitchTenant(ctx context.Context, req *v1.SwitchTenantReq) (res *v1.SwitchTenantRes, err error)
}
