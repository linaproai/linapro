// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJobGroup is the golang structure of table sys_job_group for DAO operations like Where/Data.
type SysJobGroup struct {
	g.Meta    `orm:"table:sys_job_group, do:true"`
	Id        any         // Job group ID
	Code      any         // Group code
	Name      any         // Group name
	Remark    any         // Remark
	SortOrder any         // Display order
	IsDefault any         // Default group flag: 1=yes, 0=no
	CreatedAt *gtime.Time // Creation time
	UpdatedAt *gtime.Time // Update time
	DeletedAt *gtime.Time // Deletion time
}
