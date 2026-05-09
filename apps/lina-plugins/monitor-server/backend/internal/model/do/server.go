// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Server is the golang structure of table plugin_monitor_server for DAO operations like Where/Data.
type Server struct {
	g.Meta    `orm:"table:plugin_monitor_server, do:true"`
	Id        any         // Record ID
	NodeName  any         // Node name (hostname)
	NodeIp    any         // Node IP address
	Data      any         // Monitoring data in structured text format, including CPU, memory, disk, network, Go runtime, and other metrics
	CreatedAt *gtime.Time // Collection time
	UpdatedAt *gtime.Time // Update time
}
