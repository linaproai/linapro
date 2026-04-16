package v1

import "github.com/gogf/gf/v2/frame/g"

// UninstallReq is the request for uninstalling a plugin.
type UninstallReq struct {
	g.Meta           `path:"/plugins/{id}" method:"delete" tags:"插件管理" summary:"卸载插件" permission:"plugin:uninstall" dc:"执行插件的卸载生命周期。源码插件与动态插件都会在此阶段先停用插件，并可按确认选项决定是否同时清理插件自有存储数据；勾选后宿主会执行 manifest/sql/uninstall 下的卸载 SQL，动态插件还会按已授权 storage paths 清理插件自有存储文件"`
	Id               string `json:"id" v:"required|length:1,64" dc:"插件唯一标识" eg:"plugin-demo-source"`
	PurgeStorageData *int   `json:"purgeStorageData" dc:"是否在卸载插件时同时清除插件自有存储数据：1=清除数据表数据与关联文件 0=保留；不传时默认清除" eg:"1"`
}

// UninstallRes is the response for uninstalling a plugin.
type UninstallRes struct {
	Id        string `json:"id" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Installed int    `json:"installed" dc:"安装状态：1=已安装 0=未安装" eg:"0"`
	Enabled   int    `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"0"`
}
