// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// HgDeviceNode is the golang structure for table hg_device_node.
type HgDeviceNode struct {
	DeviceId string `json:"deviceId" orm:"device_id" description:"设备国标ID（对应device_code）"`
	NodeNum  int    `json:"nodeNum"  orm:"node_num"  description:"节点编号"`
}
