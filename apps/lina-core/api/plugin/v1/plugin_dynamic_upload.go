package v1

import "github.com/gogf/gf/v2/frame/g"

// UploadDynamicPackageReq is the request for uploading a dynamic wasm package.
type UploadDynamicPackageReq struct {
	g.Meta           `path:"/plugins/dynamic/package" method:"post" mime:"multipart/form-data" tags:"插件管理" summary:"上传动态插件包" permission:"plugin:install" dc:"上传一个 dynamic wasm 动态插件包到 plugin.dynamic.storagePath。宿主会校验 wasm 文件头、自定义节、嵌入清单、ABI 版本和可选嵌入 SQL 资源，然后将产物规范化写入 <storagePath>/<plugin-id>.wasm；当前已安装的动态插件暂不支持直接通过上传覆盖升级"`
	OverwriteSupport int `json:"overwriteSupport" dc:"是否允许覆盖尚未安装的同 ID 动态插件文件：1=允许 0=不允许，默认0" eg:"0"`
}

// UploadDynamicPackageRes is the response for uploading a dynamic wasm package.
type UploadDynamicPackageRes struct {
	Id          string `json:"id" dc:"插件唯一标识，来自 wasm 嵌入清单" eg:"plugin-dynamic-demo"`
	Name        string `json:"name" dc:"插件名称，来自 wasm 嵌入清单" eg:"Dynamic Demo Plugin"`
	Version     string `json:"version" dc:"插件版本号，来自 wasm 嵌入清单" eg:"v0.1.0"`
	Type        string `json:"type" dc:"插件一级类型，固定为 dynamic" eg:"dynamic"`
	RuntimeKind string `json:"runtimeKind" dc:"运行时产物类型，当前仅支持 wasm" eg:"wasm"`
	RuntimeAbi  string `json:"runtimeAbi" dc:"运行时产物 ABI 版本，当前校验通过后返回 v1" eg:"v1"`
	Installed   int    `json:"installed" dc:"安装状态：1=已安装 0=未安装；上传后默认仍为未安装" eg:"0"`
	Enabled     int    `json:"enabled" dc:"启用状态：1=启用 0=禁用；上传后默认仍为禁用" eg:"0"`
}
