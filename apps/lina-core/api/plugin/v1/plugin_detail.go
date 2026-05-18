// This file defines the plugin detail management API DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DetailReq is the request for querying one plugin detail.
type DetailReq struct {
	g.Meta `path:"/plugins/{id}" method:"get" tags:"Plugin Management" summary:"Get plugin detail" permission:"plugin:query" dc:"Query one plugin management detail by plugin ID, including lifecycle state, runtime upgrade state, effective version, discovered version, authorization, dependencies, declared routes, and recent upgrade failure diagnostics."`
	Id     string `json:"id" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
}

// DetailRes is the response for querying one plugin detail.
type DetailRes struct {
	PluginItem
}
