// This file declares media device strategy binding delete DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteDeviceBindingReq defines the request for deleting one device strategy binding.
type DeleteDeviceBindingReq struct {
	g.Meta   `path:"/media/device-bindings/{deviceId}" method:"delete" tags:"媒体管理" summary:"删除设备策略绑定" dc:"按设备国标ID删除设备策略绑定。" permission:"media:management:remove"`
	DeviceId string `json:"deviceId" v:"required|length:1,255#设备国标ID不能为空|设备国标ID长度不能超过255个字符" dc:"设备国标ID" eg:"34020000001320000001"`
}

// DeleteDeviceBindingRes defines the response for deleting one device strategy binding.
type DeleteDeviceBindingRes struct {
	DeviceId string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000001"`
}
