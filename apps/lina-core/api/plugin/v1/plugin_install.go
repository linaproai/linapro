package v1

import "github.com/gogf/gf/v2/frame/g"

// InstallReq is the request for installing a dynamic plugin.
type InstallReq struct {
	g.Meta        `path:"/plugins/{id}/install" method:"post" tags:"插件管理" summary:"安装动态插件" permission:"plugin:install" dc:"执行动态插件的安装生命周期，包括运行插件声明的安装SQL并将插件状态更新为已安装；若插件声明了资源型 hostServices（如 storage.resources.paths、network 的 URL 模式或 data.resources.tables），则本次请求同时提交宿主确认后的授权结果"`
	Id            string                       `json:"id" v:"required|length:1,64" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Authorization *HostServiceAuthorizationReq `json:"authorization,omitempty" dc:"宿主确认后的 hostServices 授权结果；不传时默认沿用当前 release 已确认快照，若尚未确认则默认按插件本次声明全量授权" eg:"{}"`
}

// InstallRes is the response for installing a dynamic plugin.
type InstallRes struct {
	Id        string `json:"id" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Installed int    `json:"installed" dc:"安装状态：1=已安装 0=未安装" eg:"1"`
	Enabled   int    `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"0"`
}
