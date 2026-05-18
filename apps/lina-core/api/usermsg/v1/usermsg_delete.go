package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg Delete API

// DeleteReq defines the request for deleting a user message.
type DeleteReq struct {
	g.Meta `path:"/user/message/{id}" method:"delete" tags:"User Messages" summary:"Delete a single message" dc:"Delete the specified message for the current user"`
	Id     int64 `json:"id" v:"required" dc:"Message ID" eg:"1"`
}

// DeleteRes Delete message response
type DeleteRes struct{}
