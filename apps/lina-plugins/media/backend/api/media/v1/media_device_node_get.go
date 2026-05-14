// This file declares media device-node detail DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetDeviceNodeReq defines the request for querying one device-node mapping.
type GetDeviceNodeReq struct {
	g.Meta   `path:"/media/device-nodes/{deviceId}" method:"get" tags:"媒体管理" summary:"查询设备节点详情" dc:"按设备国标ID查询设备节点关系。" permission:"media:management:query"`
	DeviceId string `json:"deviceId" v:"required|length:1,64#设备国标ID不能为空|设备国标ID长度不能超过64个字符" dc:"设备国标ID" eg:"34020000001320000001"`
}

// GetDeviceNodeRes defines the response for querying one device-node mapping.
type GetDeviceNodeRes = DeviceNodeListItem
