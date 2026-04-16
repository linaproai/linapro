package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserMsg List API

// ListReq defines the request for listing user messages.
type ListReq struct {
	g.Meta   `path:"/user/message" method:"get" tags:"用户消息" summary:"获取消息列表" dc:"分页查询当前登录用户的消息列表，包括已读和未读消息"`
	PageNum  int `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
}

// ListRes User message list response
type ListRes struct {
	List  []*MessageItem `json:"list" dc:"消息列表" eg:"[]"`
	Total int            `json:"total" dc:"总条数" eg:"20"`
}

// MessageItem defines one user message list item.
type MessageItem struct {
	Id         int64       `json:"id" dc:"消息ID" eg:"1"`
	UserId     int64       `json:"userId" dc:"接收用户ID" eg:"1"`
	Title      string      `json:"title" dc:"消息标题" eg:"系统维护通知"`
	Type       int         `json:"type" dc:"消息类型：1=通知 2=公告" eg:"1"`
	SourceType string      `json:"sourceType" dc:"来源类型：notice=通知公告 plugin=动态插件 system=系统" eg:"notice"`
	SourceId   int64       `json:"sourceId" dc:"来源ID，当前通知公告预览使用该字段" eg:"1001"`
	IsRead     int         `json:"isRead" dc:"是否已读：0=未读 1=已读" eg:"0"`
	ReadAt     *gtime.Time `json:"readAt" dc:"已读时间，未读时为空" eg:"2026-04-15 16:00:00"`
	CreatedAt  *gtime.Time `json:"createdAt" dc:"消息创建时间" eg:"2026-04-15 15:30:00"`
}
