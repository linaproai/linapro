// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysNotifyChannel is the golang structure for table sys_notify_channel.
type SysNotifyChannel struct {
	Id          int64       `json:"id"          orm:"id"           description:"Primary key ID"`
	ChannelKey  string      `json:"channelKey"  orm:"channel_key"  description:"Channel key"`
	Name        string      `json:"name"        orm:"name"         description:"Channel name"`
	ChannelType string      `json:"channelType" orm:"channel_type" description:"Channel type: inbox=in-app message, email=email, webhook=webhook"`
	Status      int         `json:"status"      orm:"status"       description:"Status: 1=enabled, 0=disabled"`
	ConfigJson  string      `json:"configJson"  orm:"config_json"  description:"Channel configuration JSON"`
	Remark      string      `json:"remark"      orm:"remark"       description:"Remark"`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:"Creation time"`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:"Update time"`
	DeletedAt   *gtime.Time `json:"deletedAt"   orm:"deleted_at"   description:"Deletion time"`
}
