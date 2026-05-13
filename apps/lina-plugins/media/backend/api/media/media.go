// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

type IMediaV1 interface {
	CreateAlias(ctx context.Context, req *v1.CreateAliasReq) (res *v1.CreateAliasRes, err error)
	DeleteAlias(ctx context.Context, req *v1.DeleteAliasReq) (res *v1.DeleteAliasRes, err error)
	GetAlias(ctx context.Context, req *v1.GetAliasReq) (res *v1.GetAliasRes, err error)
	ListAliases(ctx context.Context, req *v1.ListAliasesReq) (res *v1.ListAliasesRes, err error)
	UpdateAlias(ctx context.Context, req *v1.UpdateAliasReq) (res *v1.UpdateAliasRes, err error)
	DeleteDeviceBinding(ctx context.Context, req *v1.DeleteDeviceBindingReq) (res *v1.DeleteDeviceBindingRes, err error)
	ListDeviceBindings(ctx context.Context, req *v1.ListDeviceBindingsReq) (res *v1.ListDeviceBindingsRes, err error)
	SaveDeviceBinding(ctx context.Context, req *v1.SaveDeviceBindingReq) (res *v1.SaveDeviceBindingRes, err error)
	CreateStrategy(ctx context.Context, req *v1.CreateStrategyReq) (res *v1.CreateStrategyRes, err error)
	DeleteStrategy(ctx context.Context, req *v1.DeleteStrategyReq) (res *v1.DeleteStrategyRes, err error)
	UpdateStrategyEnable(ctx context.Context, req *v1.UpdateStrategyEnableReq) (res *v1.UpdateStrategyEnableRes, err error)
	GetStrategy(ctx context.Context, req *v1.GetStrategyReq) (res *v1.GetStrategyRes, err error)
	SetGlobalStrategy(ctx context.Context, req *v1.SetGlobalStrategyReq) (res *v1.SetGlobalStrategyRes, err error)
	ListStrategies(ctx context.Context, req *v1.ListStrategiesReq) (res *v1.ListStrategiesRes, err error)
	ResolveStrategy(ctx context.Context, req *v1.ResolveStrategyReq) (res *v1.ResolveStrategyRes, err error)
	UpdateStrategy(ctx context.Context, req *v1.UpdateStrategyReq) (res *v1.UpdateStrategyRes, err error)
	DeleteTenantBinding(ctx context.Context, req *v1.DeleteTenantBindingReq) (res *v1.DeleteTenantBindingRes, err error)
	ListTenantBindings(ctx context.Context, req *v1.ListTenantBindingsReq) (res *v1.ListTenantBindingsRes, err error)
	SaveTenantBinding(ctx context.Context, req *v1.SaveTenantBindingReq) (res *v1.SaveTenantBindingRes, err error)
	DeleteTenantDeviceBinding(ctx context.Context, req *v1.DeleteTenantDeviceBindingReq) (res *v1.DeleteTenantDeviceBindingRes, err error)
	ListTenantDeviceBindings(ctx context.Context, req *v1.ListTenantDeviceBindingsReq) (res *v1.ListTenantDeviceBindingsRes, err error)
	SaveTenantDeviceBinding(ctx context.Context, req *v1.SaveTenantDeviceBindingReq) (res *v1.SaveTenantDeviceBindingRes, err error)
	CreateTenantWhite(ctx context.Context, req *v1.CreateTenantWhiteReq) (res *v1.CreateTenantWhiteRes, err error)
	DeleteTenantWhite(ctx context.Context, req *v1.DeleteTenantWhiteReq) (res *v1.DeleteTenantWhiteRes, err error)
	GetTenantWhite(ctx context.Context, req *v1.GetTenantWhiteReq) (res *v1.GetTenantWhiteRes, err error)
	ListTenantWhites(ctx context.Context, req *v1.ListTenantWhitesReq) (res *v1.ListTenantWhitesRes, err error)
	UpdateTenantWhite(ctx context.Context, req *v1.UpdateTenantWhiteReq) (res *v1.UpdateTenantWhiteRes, err error)
}
