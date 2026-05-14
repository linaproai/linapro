// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// HgDeviceNode is the golang structure of table hg_device_node for DAO operations like Where/Data.
type HgDeviceNode struct {
	g.Meta   `orm:"table:hg_device_node, do:true"`
	DeviceId any // 设备国标ID（对应device_code）
	NodeNum  any // 节点编号
}
