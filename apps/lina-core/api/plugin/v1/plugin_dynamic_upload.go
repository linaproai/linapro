// This file defines dynamic-plugin package upload API DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UploadDynamicPackageReq is the request for uploading a dynamic wasm package.
type UploadDynamicPackageReq struct {
	g.Meta           `path:"/plugins/dynamic/package" method:"post" mime:"multipart/form-data" tags:"Plugin Management" summary:"Upload dynamic plugin package" permission:"plugin:install" dc:"Upload a dynamic wasm dynamic plugin package to plugin.dynamic.storagePath. The host will verify the wasm file header, custom section, embedded manifest, ABI version and optional embedded SQL resources, and then write the product into <storagePath>/<plugin-id>.wasm in a standardized manner; if a dynamic plugin is currently installed and a higher version is uploaded, the host keeps the current effective release unchanged and exposes the target version as a pending runtime upgrade."`
	OverwriteSupport int `json:"overwriteSupport" dc:"Whether to allow overwriting of dynamic plugin files with the same ID that have not been installed yet: 1=allowed 0=not allowed, default 0" eg:"0"`
}

// UploadDynamicPackageRes is the response for uploading a dynamic wasm package.
type UploadDynamicPackageRes struct {
	Id          string `json:"id" dc:"Plugin unique identifier, from wasm embed manifest" eg:"plugin-dynamic-demo"`
	Name        string `json:"name" dc:"Plugin name, from wasm embedding manifest" eg:"Dynamic Demo Plugin"`
	Version     string `json:"version" dc:"Plugin version number, from wasm embed manifest" eg:"v0.1.0"`
	Type        string `json:"type" dc:"Plugin first-level type, fixed to dynamic" eg:"dynamic"`
	RuntimeKind string `json:"runtimeKind" dc:"Runtime product type, currently only supports wasm" eg:"wasm"`
	RuntimeAbi  string `json:"runtimeAbi" dc:"Runtime product ABI version, v1 will be returned after the current verification passes." eg:"v1"`
	Installed   int    `json:"installed" dc:"Installation status: 1=installed 0=not installed; the default is still not installed after uploading" eg:"0"`
	Enabled     int    `json:"enabled" dc:"Enabled status: 1=enabled 0=disabled; the default is still disabled after uploading" eg:"0"`
}
