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
	CategoryCode  string      `json:"categoryCode" dc:"Opaque sender-declared inbox category code (for example notice, announcement, system, alert). Hosts and plugins register translations at i18n keys notify.category.{code}.label and notify.category.{code}.color so the preview UI stays category-agnostic" eg:"notice"`
	TypeLabel     string      `json:"typeLabel" dc:"Localized category label resolved by the host according to the request locale" eg:"Notice"`
	TypeColor     string      `json:"typeColor" dc:"Localized category tag color resolved by the host so the inbox preview can render badges without hardcoding category-specific colors" eg:"blue"`
	SourceType    string      `json:"sourceType" dc:"Source type: notice=notification announcement plugin=dynamic plugin system=system" eg:"notice"`
	SourceId      int64       `json:"sourceId" dc:"Source ID" eg:"1001"`
	Content       string      `json:"content" dc:"Can be used directly to preview rendered message content" eg:"<p>The system will be undergoing maintenance tonight</p>"`
	CreatedByName string      `json:"createdByName" dc:"Sender username" eg:"admin"`
	CreatedAt     *gtime.Time `json:"createdAt" dc:"Message creation time" eg:"2026-04-21 17:00:00"`
}
