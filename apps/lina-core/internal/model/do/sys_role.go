// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysRole is the golang structure of table sys_role for DAO operations like Where/Data.
type SysRole struct {
	g.Meta    `orm:"table:sys_role, do:true"`
	Id        any        // Role ID
	TenantId  any        // Owning tenant ID, 0 means PLATFORM
	Name      any        // Role name
	Key       any        // Permission key
	Sort      any        // Display order
	DataScope any        // Data scope: 1=all, 2=tenant, 3=department, 4=self
	Status    any        // Status: 0=disabled, 1=enabled
	Remark    any        // Remark
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
	DeletedAt *time.Time // Deletion time
}
