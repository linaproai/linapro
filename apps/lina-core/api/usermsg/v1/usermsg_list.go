package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserMsg List API

// ListReq defines the request for listing user messages.
type ListReq struct {
	g.Meta   `path:"/user/message" method:"get" tags:"User Messages" summary:"Get message list" dc:"Query the message list of the currently logged in user in paging, including read and unread messages"`
	PageNum  int `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize int `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
}

// ListRes User message list response
type ListRes struct {
	List  []*MessageItem `json:"list" dc:"Message list" eg:"[]"`
	Total int            `json:"total" dc:"Total number of items" eg:"20"`
}

// MessageItem defines one user message list item.
type MessageItem struct {
	Id         int64       `json:"id" dc:"Message ID" eg:"1"`
	UserId     int64       `json:"userId" dc:"Receive user ID" eg:"1"`
	Title      string      `json:"title" dc:"Message title" eg:"System maintenance notification"`
	Type       int         `json:"type" dc:"Message type: 1=Notification 2=Announcement" eg:"1"`
	SourceType string      `json:"sourceType" dc:"Source type: notice=notification announcement plugin=dynamic plugin system=system" eg:"notice"`
	SourceId   int64       `json:"sourceId" dc:"Source ID, this field is used for the current notification announcement preview" eg:"1001"`
	IsRead     int         `json:"isRead" dc:"Whether it has been read: 0=unread 1=read" eg:"0"`
	ReadAt     *gtime.Time `json:"readAt" dc:"Read time, empty when unread" eg:"2026-04-15 16:00:00"`
	CreatedAt  *gtime.Time `json:"createdAt" dc:"Message creation time" eg:"2026-04-15 15:30:00"`
}
