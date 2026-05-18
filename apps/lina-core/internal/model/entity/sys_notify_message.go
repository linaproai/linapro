// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysNotifyMessage is the golang structure for table sys_notify_message.
type SysNotifyMessage struct {
	Id           int64       `json:"id"           orm:"id"             description:"Primary key ID"`
	TenantId     int         `json:"tenantId"     orm:"tenant_id"      description:"Owning tenant ID, 0 means PLATFORM"`
	PluginId     string      `json:"pluginId"     orm:"plugin_id"      description:"Source plugin ID, empty for host built-in flows"`
	SourceType   string      `json:"sourceType"   orm:"source_type"    description:"Source type: notice=notice, plugin=plugin, system=system"`
	SourceId     string      `json:"sourceId"     orm:"source_id"      description:"Source business ID"`
	CategoryCode string      `json:"categoryCode" orm:"category_code"  description:"Message category: notice=notification, announcement=announcement, other=other"`
	Title        string      `json:"title"        orm:"title"          description:"Message title"`
	Content      string      `json:"content"      orm:"content"        description:"Message body"`
	PayloadJson  string      `json:"payloadJson"  orm:"payload_json"   description:"Extended payload JSON"`
	SenderUserId int64       `json:"senderUserId" orm:"sender_user_id" description:"Sender user ID"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"     description:"Creation time"`
}
