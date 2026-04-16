package v1

import "github.com/gogf/gf/v2/frame/g"

// UninstallReq is the request for uninstalling a plugin.
type UninstallReq struct {
	g.Meta `path:"/plugins/{id}" method:"delete" tags:"插件管理" summary:"卸载插件" permission:"plugin:uninstall" dc:"执行插件的卸载生命周期。源码插件会在此阶段停用插件、运行 manifest/sql/uninstall 下的卸载 SQL、清理菜单与治理资源并回写未安装状态；动态插件会继续执行运行时卸载流程"`
	Id     string `json:"id" v:"required|length:1,64" dc:"插件唯一标识" eg:"plugin-demo-source"`
}

// UninstallRes is the response for uninstalling a plugin.
type UninstallRes struct {
	Id        string `json:"id" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Installed int    `json:"installed" dc:"安装状态：1=已安装 0=未安装" eg:"0"`
	Enabled   int    `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"0"`
}
