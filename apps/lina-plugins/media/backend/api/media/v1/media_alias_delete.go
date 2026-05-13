// This file declares media stream alias delete DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteAliasReq defines the request for deleting one stream alias.
type DeleteAliasReq struct {
	g.Meta `path:"/media/stream-aliases/{id}" method:"delete" tags:"媒体管理" summary:"删除流别名" dc:"删除指定流别名。" permission:"media:management:remove"`
	Id     int64 `json:"id" v:"required|min:1" dc:"ID" eg:"1"`
}

// DeleteAliasRes defines the response for deleting one stream alias.
type DeleteAliasRes struct {
	Id int64 `json:"id" dc:"已删除流别名ID" eg:"1"`
}
