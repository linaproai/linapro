// This file declares media stream alias create DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateAliasReq defines the request for creating one stream alias.
type CreateAliasReq struct {
	g.Meta     `path:"/media/stream-aliases" method:"post" tags:"媒体管理" summary:"新增流别名" dc:"新增一条流别名到真实流路径的映射。" permission:"media:management:add"`
	Alias      string `json:"alias" v:"required|length:1,255#流别名不能为空|流别名长度不能超过255个字符" dc:"流别名" eg:"camera-01"`
	AutoRemove int    `json:"autoRemove" d:"0" dc:"是否自动移除：1是，0否" eg:"0"`
	StreamPath string `json:"streamPath" v:"required|length:1,255#真实流路径不能为空|真实流路径长度不能超过255个字符" dc:"真实流路径" eg:"live/camera-01"`
}

// CreateAliasRes defines the response for creating one stream alias.
type CreateAliasRes struct {
	Id int64 `json:"id" dc:"新建流别名ID" eg:"1"`
}
