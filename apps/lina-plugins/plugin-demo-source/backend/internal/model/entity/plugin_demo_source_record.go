// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// PluginDemoSourceRecord is the golang structure for table plugin_demo_source_record.
type PluginDemoSourceRecord struct {
	Id             int64       `json:"id"             orm:"id"              description:"Primary key ID"`
	TenantId       int         `json:"tenantId"       orm:"tenant_id"       description:"Owning tenant ID, 0 means PLATFORM"`
	Title          string      `json:"title"          orm:"title"           description:"Record title"`
	Content        string      `json:"content"        orm:"content"         description:"Record content"`
	AttachmentName string      `json:"attachmentName" orm:"attachment_name" description:"Original attachment file name"`
	AttachmentPath string      `json:"attachmentPath" orm:"attachment_path" description:"Relative attachment storage path"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"      description:"Creation time"`
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"      description:"Update time"`
}
