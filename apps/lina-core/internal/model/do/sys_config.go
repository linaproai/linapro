// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysConfig is the golang structure of table sys_config for DAO operations like Where/Data.
type SysConfig struct {
	g.Meta           `orm:"table:sys_config, do:true"`
	Id               any        // Config parameter ID
	TenantId         any        // Owning tenant ID, 0 means PLATFORM default
	Name             any        // Config parameter name
	Key              any        // Config parameter key
	Value            any        // Config parameter value
	IsBuiltin        any        // Built-in record flag: 1=yes, 0=no
	Remark           any        // Remark
	CreatedAt        *time.Time // Creation time
	UpdatedAt        *time.Time // Modification time
	DeletedAt        *time.Time // Deletion time
	ValueType        any        // Parameter value input type: text, textarea, number, boolean, select, radio, multi_select, richtext
	Options          any        // JSON array of {label,value} options for select/radio/multi_select; empty for other types
	SystemManageable any        // Whether the parameter may be governed on the system parameter admin surface: 1=yes, 0=no
}
