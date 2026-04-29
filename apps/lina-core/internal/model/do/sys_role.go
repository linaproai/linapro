// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysRole is the golang structure of table sys_role for DAO operations like Where/Data.
type SysRole struct {
	g.Meta    `orm:"table:sys_role, do:true"`
	Id        any         // Role ID
	Name      any         // Role name
	Key       any         // Permission key
	Sort      any         // Display order
	DataScope any         // Data scope: 1=all, 2=department, 3=self
	Status    any         // Status: 0=disabled, 1=enabled
	Remark    any         // Remark
	CreatedAt *gtime.Time // Creation time
	UpdatedAt *gtime.Time // Update time
	DeletedAt *gtime.Time // Deletion time
}
