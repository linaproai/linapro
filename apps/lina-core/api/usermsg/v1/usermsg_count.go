package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg Count API

// CountReq defines the request for retrieving unread message counts.
type CountReq struct {
	g.Meta `path:"/user/message/count" method:"get" tags:"用户消息" summary:"获取未读消息数量" dc:"获取当前登录用户的未读消息数量，用于前端消息图标角标展示"`
}

// CountRes Unread message count response
type CountRes struct {
	Count int `json:"count" dc:"未读消息数量" eg:"5"`
}
