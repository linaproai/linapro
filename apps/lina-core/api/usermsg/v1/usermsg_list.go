// This file defines user-message list DTOs and message source enum values.

package v1

import (
	"lina-core/pkg/statusflag"

	"github.com/gogf/gf/v2/frame/g"
)

// SourceType identifies the business origin of one inbox message.
type SourceType string

const (
	// SourceTypeNotice identifies notice-originated messages.
	SourceTypeNotice SourceType = "notice"
	// SourceTypePlugin identifies plugin-originated messages.
	SourceTypePlugin SourceType = "plugin"
	// SourceTypeSystem identifies system-originated messages.
	SourceTypeSystem SourceType = "system"
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
	Id           int64                `json:"id" dc:"Message ID" eg:"1"`
	UserId       int64                `json:"userId" dc:"Receive user ID" eg:"1"`
	Title        string               `json:"title" dc:"Message title" eg:"System maintenance notification"`
	CategoryCode string               `json:"categoryCode" dc:"Opaque sender-declared inbox category code (for example notice, announcement, system, alert). Hosts and plugins register translations at i18n keys notify.category.{code}.label and notify.category.{code}.color so the inbox UI stays category-agnostic" eg:"notice"`
	TypeLabel    string               `json:"typeLabel" dc:"Localized category label resolved by the host according to the request locale" eg:"Notice"`
	TypeColor    string               `json:"typeColor" dc:"Localized category tag color resolved by the host so the inbox UI can render badges without hardcoding category-specific colors" eg:"blue"`
	SourceType   SourceType           `json:"sourceType" dc:"Source type: notice=notification announcement plugin=dynamic plugin system=system" eg:"notice"`
	SourceId     string               `json:"sourceId" dc:"Original sender-declared source record ID" eg:"notice-1001"`
	IsRead       statusflag.ReadState `json:"isRead" dc:"Whether it has been read: 0=unread 1=read" eg:"0"`
	ReadAt       *int64               `json:"readAt" dc:"Read time as Unix timestamp in milliseconds, empty when unread" eg:"1776240000000"`
	CreatedAt    *int64               `json:"createdAt" dc:"Message creation time as Unix timestamp in milliseconds" eg:"1776238200000"`
}
