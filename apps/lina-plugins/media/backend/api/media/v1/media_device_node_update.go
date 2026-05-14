// This file declares media device-node update DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateDeviceNodeReq defines the request for updating one device-node mapping.
type UpdateDeviceNodeReq struct {
	g.Meta      `path:"/media/device-nodes/{oldDeviceId}" method:"put" tags:"媒体管理" summary:"修改设备节点" dc:"按原设备国标ID修改设备节点关系。" permission:"media:management:edit"`
	OldDeviceId string `json:"oldDeviceId" v:"required|length:1,64#原设备国标ID不能为空|原设备国标ID长度不能超过64个字符" dc:"原设备国标ID" eg:"34020000001320000001"`
	DeviceId    string `json:"deviceId" v:"required|length:1,64#设备国标ID不能为空|设备国标ID长度不能超过64个字符" dc:"设备国标ID" eg:"34020000001320000002"`
	NodeNum     int    `json:"nodeNum" v:"min:0|max:255#节点编号不能小于0|节点编号不能大于255" dc:"节点编号" eg:"1"`
}

// UpdateDeviceNodeRes defines the response for updating one device-node mapping.
type UpdateDeviceNodeRes struct {
	DeviceId string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000002"`
}
