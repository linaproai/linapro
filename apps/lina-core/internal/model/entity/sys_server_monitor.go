// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysServerMonitor is the golang structure for table sys_server_monitor.
type SysServerMonitor struct {
	Id        int64       `json:"id"        orm:"id"         description:"记录ID"`
	NodeName  string      `json:"nodeName"  orm:"node_name"  description:"节点名称（hostname）"`
	NodeIp    string      `json:"nodeIp"    orm:"node_ip"    description:"节点IP地址"`
	Data      string      `json:"data"      orm:"data"       description:"监控数据（JSON格式，包含CPU、内存、磁盘、网络、Go运行时等指标）"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"采集时间"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`
}
