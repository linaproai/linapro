package v1

import "github.com/gogf/gf/v2/frame/g"

// DisableReq is the request for disabling plugin.
type DisableReq struct {
	g.Meta `path:"/plugins/{id}/disable" method:"put" tags:"Plugin Management" summary:"Disable plugin" permission:"plugin:disable" dc:"Mark the specified plugin as disabled and write the plugin status configuration"`
	Id     string `json:"id" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
}

// DisableRes is the response for disabling plugin.
type DisableRes struct {
	Id      string `json:"id" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	Enabled int    `json:"enabled" dc:"Enabled status: 1=enabled 0=disabled" eg:"0"`
}
