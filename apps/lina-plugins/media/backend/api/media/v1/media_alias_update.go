// This file declares media stream alias update DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateAliasReq defines the request for updating one stream alias.
type UpdateAliasReq struct {
	g.Meta     `path:"/media/stream-aliases/{id}" method:"put" tags:"媒体管理" summary:"修改流别名" dc:"修改指定流别名的别名、真实流路径和自动移除标记。" permission:"media:management:edit"`
	Id         int64  `json:"id" v:"required|min:1" dc:"ID" eg:"1"`
	Alias      string `json:"alias" v:"required|length:1,255#流别名不能为空|流别名长度不能超过255个字符" dc:"流别名" eg:"camera-01"`
	AutoRemove int    `json:"autoRemove" d:"0" dc:"是否自动移除：1是，0否" eg:"0"`
	StreamPath string `json:"streamPath" v:"required|length:1,255#真实流路径不能为空|真实流路径长度不能超过255个字符" dc:"真实流路径" eg:"live/camera-01"`
}

// UpdateAliasRes defines the response for updating one stream alias.
type UpdateAliasRes struct {
	Id int64 `json:"id" dc:"已更新流别名ID" eg:"1"`
}
