// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaStreamAlias is the golang structure for table media_stream_alias.
type MediaStreamAlias struct {
	Id         int64       `json:"id"         orm:"id"          description:"ID"`
	Alias      string      `json:"alias"      orm:"alias"       description:"流别名"`
	AutoRemove int         `json:"autoRemove" orm:"auto_remove" description:"是否自动移除：1是，0否"`
	StreamPath string      `json:"streamPath" orm:"stream_path" description:"真实流路径"`
	CreateTime *gtime.Time `json:"createTime" orm:"create_time" description:"创建时间"`
}
