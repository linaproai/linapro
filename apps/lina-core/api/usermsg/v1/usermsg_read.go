package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg Read API

// ReadReq defines the request for marking a user message as read.
type ReadReq struct {
	g.Meta `path:"/user/message/{id}/read" method:"put" tags:"用户消息" summary:"标记消息已读" dc:"将指定消息标记为已读状态"`
	Id     int64 `json:"id" v:"required" dc:"消息ID" eg:"1"`
}

// ReadRes Mark message as read response
type ReadRes struct{}
