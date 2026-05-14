// This file declares media device-node create DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateDeviceNodeReq defines the request for creating one device-node mapping.
type CreateDeviceNodeReq struct {
	g.Meta   `path:"/media/device-nodes" method:"post" tags:"媒体管理" summary:"新增设备节点" dc:"新增一条设备节点关系。" permission:"media:management:add"`
	DeviceId string `json:"deviceId" v:"required|length:1,64#设备国标ID不能为空|设备国标ID长度不能超过64个字符" dc:"设备国标ID" eg:"34020000001320000001"`
	NodeNum  int    `json:"nodeNum" v:"min:0|max:255#节点编号不能小于0|节点编号不能大于255" dc:"节点编号" eg:"1"`
}

// CreateDeviceNodeRes defines the response for creating one device-node mapping.
type CreateDeviceNodeRes struct {
	DeviceId string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000001"`
}
