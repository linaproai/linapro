package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg ReadAll API

// ReadAllReq defines the request for marking all user messages as read.
type ReadAllReq struct {
	g.Meta `path:"/user/message/read-all" method:"put" tags:"用户消息" summary:"标记全部消息已读" dc:"将当前用户的所有未读消息批量标记为已读"`
}

// ReadAllRes Mark all messages as read response
type ReadAllRes struct{}
