package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg Clear API

// ClearReq defines the request for clearing all user messages.
type ClearReq struct {
	g.Meta `path:"/user/message/clear" method:"delete" tags:"用户消息" summary:"清空全部消息" dc:"清空当前用户的所有消息记录，包括已读和未读"`
}

// ClearRes Clear messages response
type ClearRes struct{}
