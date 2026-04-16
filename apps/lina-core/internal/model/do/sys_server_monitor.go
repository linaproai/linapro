// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysServerMonitor is the golang structure of table sys_server_monitor for DAO operations like Where/Data.
type SysServerMonitor struct {
	g.Meta    `orm:"table:sys_server_monitor, do:true"`
	Id        any         // 记录ID
	NodeName  any         // 节点名称（hostname）
	NodeIp    any         // 节点IP地址
	Data      any         // 监控数据（JSON格式，包含CPU、内存、磁盘、网络、Go运行时等指标）
	CreatedAt *gtime.Time // 采集时间
	UpdatedAt *gtime.Time // 更新时间
}
