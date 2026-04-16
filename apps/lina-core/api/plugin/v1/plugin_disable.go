package v1

import "github.com/gogf/gf/v2/frame/g"

// DisableReq is the request for disabling plugin.
type DisableReq struct {
	g.Meta `path:"/plugins/{id}/disable" method:"put" tags:"插件管理" summary:"禁用插件" permission:"plugin:disable" dc:"将指定插件标记为禁用状态，并写入插件状态配置"`
	Id     string `json:"id" v:"required|length:1,64" dc:"插件唯一标识" eg:"plugin-demo-source"`
}

// DisableRes is the response for disabling plugin.
type DisableRes struct {
	Id      string `json:"id" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Enabled int    `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"0"`
}
