// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysMenu is the golang structure of table sys_menu for DAO operations like Where/Data.
type SysMenu struct {
	g.Meta     `orm:"table:sys_menu, do:true"`
	Id         any         // Menu ID
	ParentId   any         // Parent menu ID, 0 means root menu
	MenuKey    any         // Stable menu business key
	Name       any         // Menu name with i18n support
	Path       any         // Route path
	Component  any         // Component path
	Perms      any         // Permission identifier
	Icon       any         // Menu icon
	Type       any         // Menu type: D=directory, M=menu, B=button
	Sort       any         // Display order
	Visible    any         // Visibility: 1=visible, 0=hidden
	Status     any         // Status: 0=disabled, 1=enabled
	IsFrame    any         // External link flag: 1=yes, 0=no
	IsCache    any         // Cache flag: 1=yes, 0=no
	QueryParam any         // Route parameters in JSON format
	Remark     any         // Remark
	CreatedAt  *gtime.Time // Creation time
	UpdatedAt  *gtime.Time // Update time
	DeletedAt  *gtime.Time // Deletion time
}
