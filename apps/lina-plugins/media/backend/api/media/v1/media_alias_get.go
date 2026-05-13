// This file declares media stream alias detail DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetAliasReq defines the request for querying one stream alias.
type GetAliasReq struct {
	g.Meta `path:"/media/stream-aliases/{id}" method:"get" tags:"媒体管理" summary:"获取流别名详情" dc:"根据ID获取流别名详情。" permission:"media:management:query"`
	Id     int64 `json:"id" v:"required|min:1" dc:"ID" eg:"1"`
}

// GetAliasRes defines one stream alias detail.
type GetAliasRes struct {
	Id         int64  `json:"id" dc:"ID" eg:"1"`
	Alias      string `json:"alias" dc:"流别名" eg:"camera-01"`
	AutoRemove int    `json:"autoRemove" dc:"是否自动移除：1是，0否" eg:"0"`
	StreamPath string `json:"streamPath" dc:"真实流路径" eg:"live/camera-01"`
	CreateTime  string `json:"createTime" dc:"创建时间" eg:"2026-05-13 10:00:00"`
}
