// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Server is the golang structure for table server.
type Server struct {
	Id        int64       `json:"id"        orm:"id"         description:"Record ID"`
	NodeName  string      `json:"nodeName"  orm:"node_name"  description:"Node name (hostname)"`
	NodeIp    string      `json:"nodeIp"    orm:"node_ip"    description:"Node IP address"`
	Data      string      `json:"data"      orm:"data"       description:"Monitoring data in structured text format, including CPU, memory, disk, network, Go runtime, and other metrics"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Collection time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
}
