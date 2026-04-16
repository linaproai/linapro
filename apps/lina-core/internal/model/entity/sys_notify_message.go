// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysNotifyMessage is the golang structure for table sys_notify_message.
type SysNotifyMessage struct {
	Id           int64       `json:"id"           orm:"id"             description:"主键ID"`
	PluginId     string      `json:"pluginId"     orm:"plugin_id"      description:"来源插件ID，宿主内建流程为空"`
	SourceType   string      `json:"sourceType"   orm:"source_type"    description:"来源类型：notice=公告 plugin=插件 system=系统"`
	SourceId     string      `json:"sourceId"     orm:"source_id"      description:"来源业务ID"`
	CategoryCode string      `json:"categoryCode" orm:"category_code"  description:"消息分类：notice=通知 announcement=公告 other=其他"`
	Title        string      `json:"title"        orm:"title"          description:"消息标题"`
	Content      string      `json:"content"      orm:"content"        description:"消息正文"`
	PayloadJson  string      `json:"payloadJson"  orm:"payload_json"   description:"扩展载荷JSON"`
	SenderUserId int64       `json:"senderUserId" orm:"sender_user_id" description:"发送者用户ID"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"     description:"创建时间"`
}
