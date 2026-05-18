// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDictType is the golang structure of table sys_dict_type for DAO operations like Where/Data.
type SysDictType struct {
	g.Meta              `orm:"table:sys_dict_type, do:true"`
	Id                  any         // Dictionary type ID
	TenantId            any         // Owning tenant ID, 0 means PLATFORM default
	Name                any         // Dictionary name
	Type                any         // Dictionary type
	Status              any         // Status: 0=disabled, 1=enabled
	IsBuiltin           any         // Built-in record flag: 1=yes, 0=no
	AllowTenantOverride any         // Whether tenants may override this dictionary type
	Remark              any         // Remark
	CreatedAt           *gtime.Time // Creation time
	UpdatedAt           *gtime.Time // Update time
	DeletedAt           *gtime.Time // Deletion time
}
