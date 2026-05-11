// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Dept is the golang structure of table plugin_org_center_dept for DAO operations like Where/Data.
type Dept struct {
	g.Meta    `orm:"table:plugin_org_center_dept, do:true"`
	Id        any         // Department ID
	TenantId  any         // Owning tenant ID, 0 means PLATFORM
	ParentId  any         // Parent department ID
	Ancestors any         // Ancestor list
	Name      any         // Department name
	Code      any         // Department code
	OrderNum  any         // Display order
	Leader    any         // Leader user ID
	Phone     any         // Contact phone number
	Email     any         // Email address
	Status    any         // Status: 0=disabled, 1=enabled
	Remark    any         // Remark
	CreatedAt *gtime.Time // Creation time
	UpdatedAt *gtime.Time // Update time
	DeletedAt *gtime.Time // Deletion time
}
