// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Post is the golang structure of table plugin_org_center_post for DAO operations like Where/Data.
type Post struct {
	g.Meta    `orm:"table:plugin_org_center_post, do:true"`
	Id        any         // Post ID
	TenantId  any         // Owning tenant ID, 0 means PLATFORM
	DeptId    any         // Owning department ID
	Code      any         // Post code
	Name      any         // Post name
	Sort      any         // Display order
	Status    any         // Status: 0=disabled, 1=enabled
	Remark    any         // Remark
	CreatedAt *gtime.Time // Creation time
	UpdatedAt *gtime.Time // Update time
	DeletedAt *gtime.Time // Deletion time
}
