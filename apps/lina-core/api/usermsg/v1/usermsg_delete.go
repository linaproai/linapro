package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg Delete API

// DeleteReq defines the request for deleting a user message.
type DeleteReq struct {
	g.Meta `path:"/user/message/{id}" method:"delete" tags:"用户消息" summary:"删除单条消息" dc:"删除当前用户的指定消息"`
	Id     int64 `json:"id" v:"required" dc:"消息ID" eg:"1"`
}

// DeleteRes Delete message response
type DeleteRes struct{}
