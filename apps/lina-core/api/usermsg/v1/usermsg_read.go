package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg Read API

// ReadReq defines the request for marking a user message as read.
type ReadReq struct {
	g.Meta `path:"/user/message/{id}/read" method:"put" tags:"User Messages" summary:"Mark message as read" dc:"Mark the specified message as read"`
	Id     int64 `json:"id" v:"required" dc:"Message ID" eg:"1"`
}

// ReadRes Mark message as read response
type ReadRes struct{}
