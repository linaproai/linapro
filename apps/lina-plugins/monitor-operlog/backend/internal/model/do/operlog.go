// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Operlog is the golang structure of table plugin_monitor_operlog for DAO operations like Where/Data.
type Operlog struct {
	g.Meta             `orm:"table:plugin_monitor_operlog, do:true"`
	Id                 any         // Log ID
	TenantId           any         // Owning tenant ID, 0 means PLATFORM
	ActingUserId       any         // Actual acting user ID for platform operations or impersonation
	OnBehalfOfTenantId any         // Target tenant ID when a platform administrator acts on behalf of a tenant
	IsImpersonation    any         // Whether this log was produced during tenant impersonation
	Title              any         // Module title
	OperSummary        any         // Operation summary
	RouteOwner         any         // Route owner: core or plugin ID
	RouteMethod        any         // Route request method
	RoutePath          any         // Route path
	RouteDocKey        any         // API documentation structured key
	OperType           any         // Operation type: create=create, update=update, delete=delete, export=export, import=import, other=other
	Method             any         // Method name
	RequestMethod      any         // Request method: GET/POST/PUT/DELETE
	OperName           any         // Operator
	OperUrl            any         // Request URL
	OperIp             any         // Operation IP address
	OperParam          any         // Request parameters
	JsonResult         any         // Response parameters
	Status             any         // Operation status: 0=succeeded, 1=failed
	ErrorMsg           any         // Error message
	CostTime           any         // Duration in milliseconds
	OperTime           *gtime.Time // Operation time
}
