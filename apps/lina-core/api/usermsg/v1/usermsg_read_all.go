package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg ReadAll API

// ReadAllReq defines the request for marking all user messages as read.
type ReadAllReq struct {
	g.Meta `path:"/user/message/read-all" method:"put" tags:"User Messages" summary:"Mark all messages as read" dc:"Mark all unread messages of the current user as read in batches"`
}

// ReadAllRes Mark all messages as read response
type ReadAllRes struct{}
