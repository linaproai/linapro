// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

type IPluginV1 interface {
	DependencyCheck(ctx context.Context, req *v1.DependencyCheckReq) (res *v1.DependencyCheckRes, err error)
	Disable(ctx context.Context, req *v1.DisableReq) (res *v1.DisableRes, err error)
	DynamicList(ctx context.Context, req *v1.DynamicListReq) (res *v1.DynamicListRes, err error)
	UploadDynamicPackage(ctx context.Context, req *v1.UploadDynamicPackageReq) (res *v1.UploadDynamicPackageRes, err error)
	Enable(ctx context.Context, req *v1.EnableReq) (res *v1.EnableRes, err error)
	Install(ctx context.Context, req *v1.InstallReq) (res *v1.InstallRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	ResourceList(ctx context.Context, req *v1.ResourceListReq) (res *v1.ResourceListRes, err error)
	Sync(ctx context.Context, req *v1.SyncReq) (res *v1.SyncRes, err error)
	UpdateTenantProvisioningPolicy(ctx context.Context, req *v1.UpdateTenantProvisioningPolicyReq) (res *v1.UpdateTenantProvisioningPolicyRes, err error)
	Uninstall(ctx context.Context, req *v1.UninstallReq) (res *v1.UninstallRes, err error)
}
