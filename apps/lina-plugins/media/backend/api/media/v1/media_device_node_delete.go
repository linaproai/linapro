// This file declares media device-node delete DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteDeviceNodeReq defines the request for deleting one device-node mapping.
type DeleteDeviceNodeReq struct {
	g.Meta   `path:"/media/device-nodes/{deviceId}" method:"delete" tags:"媒体管理" summary:"删除设备节点" dc:"按设备国标ID删除设备节点关系。" permission:"media:management:remove"`
	DeviceId string `json:"deviceId" v:"required|length:1,64#设备国标ID不能为空|设备国标ID长度不能超过64个字符" dc:"设备国标ID" eg:"34020000001320000001"`
}

// DeleteDeviceNodeRes defines the response for deleting one device-node mapping.
type DeleteDeviceNodeRes struct {
	DeviceId string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000001"`
}
