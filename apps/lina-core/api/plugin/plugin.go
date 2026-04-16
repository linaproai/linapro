// Package plugin defines plugin API route contracts for plugin management and
// public runtime state queries.
package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

// IPluginV1 defines plugin management APIs.
type IPluginV1 interface {
	// DynamicList returns public dynamic-plugin states for slot rendering.
	DynamicList(ctx context.Context, req *v1.DynamicListReq) (res *v1.DynamicListRes, err error)
	// List returns discovered plugins and synchronized status.
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	// Sync scans source plugin manifests and synchronizes registry metadata.
	Sync(ctx context.Context, req *v1.SyncReq) (res *v1.SyncRes, err error)
	// Install executes dynamic plugin install lifecycle.
	Install(ctx context.Context, req *v1.InstallReq) (res *v1.InstallRes, err error)
	// UploadDynamicPackage uploads one dynamic wasm package into the plugin workspace.
	UploadDynamicPackage(ctx context.Context, req *v1.UploadDynamicPackageReq) (res *v1.UploadDynamicPackageRes, err error)
	// Enable sets a plugin to enabled status.
	Enable(ctx context.Context, req *v1.EnableReq) (res *v1.EnableRes, err error)
	// Disable sets a plugin to disabled status.
	Disable(ctx context.Context, req *v1.DisableReq) (res *v1.DisableRes, err error)
	// Uninstall executes dynamic plugin uninstall lifecycle.
	Uninstall(ctx context.Context, req *v1.UninstallReq) (res *v1.UninstallRes, err error)
	// ResourceList queries plugin-owned backend resources through the generic plugin resource contract.
	ResourceList(ctx context.Context, req *v1.ResourceListReq) (res *v1.ResourceListRes, err error)
}
