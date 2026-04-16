package v1

import "github.com/gogf/gf/v2/frame/g"

// UninstallReq is the request for uninstalling a dynamic plugin.
type UninstallReq struct {
	g.Meta `path:"/plugins/{id}" method:"delete" tags:"插件管理" summary:"卸载动态插件" permission:"plugin:uninstall" dc:"执行动态插件的卸载生命周期，包括停用插件、运行插件声明的卸载SQL并将插件状态更新为未安装；源码插件随宿主编译集成，不支持调用该接口卸载"`
	Id     string `json:"id" v:"required|length:1,64" dc:"插件唯一标识" eg:"plugin-demo-source"`
}

// UninstallRes is the response for uninstalling a dynamic plugin.
type UninstallRes struct {
	Id        string `json:"id" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Installed int    `json:"installed" dc:"安装状态：1=已安装 0=未安装" eg:"0"`
	Enabled   int    `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"0"`
}
