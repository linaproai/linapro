package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserMsg Get API

// GetReq defines the request for retrieving one current-user message detail.
type GetReq struct {
	g.Meta `path:"/user/message/{id}" method:"get" tags:"User Messages" summary:"Get message details" dc:"Get the details of a message of the currently logged in user and return the title, content, message type and sender information required to preview the pop-up window."`
	Id     int64 `json:"id" v:"required" dc:"Message ID" eg:"1"`
}

// GetRes defines the response for retrieving one current-user message detail.
type GetRes struct {
	Id            int64       `json:"id" dc:"Message ID" eg:"1"`
	Title         string      `json:"title" dc:"Message title" eg:"System maintenance notification"`
	Type          int         `json:"type" dc:"Message type: 1=Notification 2=Announcement" eg:"1"`
	SourceType    string      `json:"sourceType" dc:"Source type: notice=notification announcement plugin=dynamic plugin system=system" eg:"notice"`
	SourceId      int64       `json:"sourceId" dc:"Source ID" eg:"1001"`
	Content       string      `json:"content" dc:"Can be used directly to preview rendered message content" eg:"<p>The system will be undergoing maintenance tonight</p>"`
	CreatedByName string      `json:"createdByName" dc:"Sender username" eg:"admin"`
	CreatedAt     *gtime.Time `json:"createdAt" dc:"Message creation time" eg:"2026-04-21 17:00:00"`
}
