// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// PluginDemoSourceRecord is the golang structure of table plugin_demo_source_record for DAO operations like Where/Data.
type PluginDemoSourceRecord struct {
	g.Meta         `orm:"table:plugin_demo_source_record, do:true"`
	Id             any         // Primary key ID
	TenantId       any         // Owning tenant ID, 0 means PLATFORM
	Title          any         // Record title
	Content        any         // Record content
	AttachmentName any         // Original attachment file name
	AttachmentPath any         // Relative attachment storage path
	CreatedAt      *gtime.Time // Creation time
	UpdatedAt      *gtime.Time // Update time
}
