// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDictData is the golang structure of table sys_dict_data for DAO operations like Where/Data.
type SysDictData struct {
	g.Meta    `orm:"table:sys_dict_data, do:true"`
	Id        any         // Dictionary data ID
	DictType  any         // Dictionary type
	Label     any         // Dictionary label
	Value     any         // Dictionary value
	Sort      any         // Display order
	TagStyle  any         // Tag style: primary/success/danger/warning, etc.
	CssClass  any         // CSS class name
	Status    any         // Status: 0=disabled, 1=enabled
	IsBuiltin any         // Built-in record flag: 1=yes, 0=no
	Remark    any         // Remark
	CreatedAt *gtime.Time // Creation time
	UpdatedAt *gtime.Time // Update time
	DeletedAt *gtime.Time // Deletion time
	TenantId  any         // Owning tenant ID, 0 means PLATFORM default
}
