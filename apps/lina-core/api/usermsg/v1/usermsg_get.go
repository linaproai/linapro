package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserMsg Get API

// GetReq defines the request for retrieving one current-user message detail.
type GetReq struct {
	g.Meta `path:"/user/message/{id}" method:"get" tags:"用户消息" summary:"获取消息详情" dc:"获取当前登录用户的一条消息详情，返回预览弹窗所需的标题、内容、消息类型和发送者信息"`
	Id     int64 `json:"id" v:"required" dc:"消息ID" eg:"1"`
}

// GetRes defines the response for retrieving one current-user message detail.
type GetRes struct {
	Id            int64       `json:"id" dc:"消息ID" eg:"1"`
	Title         string      `json:"title" dc:"消息标题" eg:"系统维护通知"`
	Type          int         `json:"type" dc:"消息类型：1=通知 2=公告" eg:"1"`
	SourceType    string      `json:"sourceType" dc:"来源类型：notice=通知公告 plugin=动态插件 system=系统" eg:"notice"`
	SourceId      int64       `json:"sourceId" dc:"来源ID" eg:"1001"`
	Content       string      `json:"content" dc:"可直接用于预览渲染的消息内容" eg:"<p>系统将在今晚进行维护</p>"`
	CreatedByName string      `json:"createdByName" dc:"发送者用户名" eg:"admin"`
	CreatedAt     *gtime.Time `json:"createdAt" dc:"消息创建时间" eg:"2026-04-21 17:00:00"`
}
