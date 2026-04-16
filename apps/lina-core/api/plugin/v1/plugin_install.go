package v1

import "github.com/gogf/gf/v2/frame/g"

// InstallReq is the request for installing a plugin.
type InstallReq struct {
	g.Meta        `path:"/plugins/{id}/install" method:"post" tags:"插件管理" summary:"安装插件" permission:"plugin:install" dc:"执行插件的安装生命周期。源码插件会在此阶段运行其 manifest/sql 安装 SQL、同步菜单与治理资源并写入已安装状态；动态插件会继续执行运行时安装流程。若目标为动态插件且声明了资源型 hostServices（如 storage.resources.paths、network 的 URL 模式或 data.resources.tables），则本次请求同时提交宿主确认后的授权结果"`
	Id            string                       `json:"id" v:"required|length:1,64" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Authorization *HostServiceAuthorizationReq `json:"authorization,omitempty" dc:"宿主确认后的 hostServices 授权结果；不传时默认沿用当前 release 已确认快照，若尚未确认则默认按插件本次声明全量授权" eg:"{}"`
}

// InstallRes is the response for installing a plugin.
type InstallRes struct {
	Id        string `json:"id" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Installed int    `json:"installed" dc:"安装状态：1=已安装 0=未安装" eg:"1"`
	Enabled   int    `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"0"`
}
