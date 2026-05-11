// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserDept is the golang structure of table plugin_org_center_user_dept for DAO operations like Where/Data.
type UserDept struct {
	g.Meta   `orm:"table:plugin_org_center_user_dept, do:true"`
	TenantId any // Owning tenant ID, 0 means PLATFORM
	UserId   any // User ID
	DeptId   any // Department ID
}
