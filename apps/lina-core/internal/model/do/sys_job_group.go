// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysJobGroup is the golang structure of table sys_job_group for DAO operations like Where/Data.
type SysJobGroup struct {
	g.Meta    `orm:"table:sys_job_group, do:true"`
	Id        any        // Job group ID
	TenantId  any        // Owning tenant ID, 0 means PLATFORM
	Code      any        // Group code
	Name      any        // Group name
	Remark    any        // Remark
	SortOrder any        // Display order
	IsDefault any        // Default group flag: 1=yes, 0=no
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
	DeletedAt *time.Time // Deletion time
}
