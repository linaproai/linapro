package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg Clear API

// ClearReq defines the request for clearing all user messages.
type ClearReq struct {
	g.Meta `path:"/user/message/clear" method:"delete" tags:"User Messages" summary:"Clear all messages" dc:"Clear all message records of the current user, including read and unread"`
}

// ClearRes Clear messages response
type ClearRes struct{}
