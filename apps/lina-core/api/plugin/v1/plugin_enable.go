package v1

import "github.com/gogf/gf/v2/frame/g"

// EnableReq is the request for enabling plugin.
type EnableReq struct {
	g.Meta        `path:"/plugins/{id}/enable" method:"put" tags:"插件管理" summary:"启用插件" permission:"plugin:enable" dc:"将指定插件标记为启用状态，并写入插件状态配置；若插件声明了资源型 hostServices（如 storage.resources.paths、network 的 URL 模式或 data.resources.tables），则本次请求同时提交宿主确认后的授权结果"`
	Id            string                       `json:"id" v:"required|length:1,64" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Authorization *HostServiceAuthorizationReq `json:"authorization,omitempty" dc:"宿主确认后的 hostServices 授权结果；不传时默认沿用当前 release 已确认快照，若尚未确认则默认按插件本次声明全量授权" eg:"{}"`
}

// EnableRes is the response for enabling plugin.
type EnableRes struct {
	Id      string `json:"id" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Enabled int    `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"1"`
}
