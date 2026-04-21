// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// PluginDemoSourceRecord is the golang structure for table plugin_demo_source_record.
type PluginDemoSourceRecord struct {
	Id             int64       `json:"id"             orm:"id"              description:"主键ID"`
	Title          string      `json:"title"          orm:"title"           description:"记录标题"`
	Content        string      `json:"content"        orm:"content"         description:"记录内容"`
	AttachmentName string      `json:"attachmentName" orm:"attachment_name" description:"附件原始文件名"`
	AttachmentPath string      `json:"attachmentPath" orm:"attachment_path" description:"附件相对存储路径"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"      description:"创建时间"`
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"      description:"更新时间"`
}
