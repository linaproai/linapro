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
	Id             any         // 主键ID
	Title          any         // 记录标题
	Content        any         // 记录内容
	AttachmentName any         // 附件原始文件名
	AttachmentPath any         // 附件相对存储路径
	CreatedAt      *gtime.Time // 创建时间
	UpdatedAt      *gtime.Time // 更新时间
}
