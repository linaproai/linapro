// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Loginlog is the golang structure of table plugin_monitor_loginlog for DAO operations like Where/Data.
type Loginlog struct {
	g.Meta             `orm:"table:plugin_monitor_loginlog, do:true"`
	Id                 any         // Log ID
	TenantId           any         // Owning tenant ID, 0 means PLATFORM
	ActingUserId       any         // Actual acting user ID for platform operations or impersonation
	OnBehalfOfTenantId any         // Target tenant ID when a platform administrator acts on behalf of a tenant
	IsImpersonation    any         // Whether this log was produced during tenant impersonation
	UserName           any         // Login account
	Status             any         // Login status: 0=succeeded, 1=failed
	Ip                 any         // Login IP address
	Browser            any         // Browser type
	Os                 any         // Operating system
	Msg                any         // Prompt message
	LoginTime          *gtime.Time // Login time
}
