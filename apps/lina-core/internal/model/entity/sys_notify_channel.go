// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysNotifyChannel is the golang structure for table sys_notify_channel.
type SysNotifyChannel struct {
	Id          int64      `json:"id"          orm:"id"           description:"Primary key ID"`
	ChannelKey  string     `json:"channelKey"  orm:"channel_key"  description:"Channel key"`
	Name        string     `json:"name"        orm:"name"         description:"Channel name"`
	ChannelType string     `json:"channelType" orm:"channel_type" description:"Channel type: inbox=in-app message, email=email, webhook=webhook"`
	Status      int        `json:"status"      orm:"status"       description:"Status: 1=enabled, 0=disabled"`
	ConfigJson  string     `json:"configJson"  orm:"config_json"  description:"Channel configuration JSON"`
	Remark      string     `json:"remark"      orm:"remark"       description:"Remark"`
	CreatedAt   *time.Time `json:"createdAt"   orm:"created_at"   description:"Creation time"`
	UpdatedAt   *time.Time `json:"updatedAt"   orm:"updated_at"   description:"Update time"`
	DeletedAt   *time.Time `json:"deletedAt"   orm:"deleted_at"   description:"Deletion time"`
}
