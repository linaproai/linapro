// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MediaDeviceNode is the golang structure for table media_device_node.
type MediaDeviceNode struct {
	DeviceId string `json:"deviceId" orm:"device_id" description:"设备国标ID（对应device_code）"`
	NodeNum  int    `json:"nodeNum"  orm:"node_num"  description:"节点编号"`
}
