// This file declares media tenant-device strategy binding delete DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteTenantDeviceBindingReq defines the request for deleting one tenant-device strategy binding.
type DeleteTenantDeviceBindingReq struct {
	g.Meta   `path:"/media/tenant-device-bindings/{tenantId}/{deviceId}" method:"delete" tags:"媒体管理" summary:"删除租户设备策略绑定" dc:"按租户ID和设备国标ID删除租户设备策略绑定。" permission:"media:management:remove"`
	TenantId string `json:"tenantId" v:"required|length:1,255#租户ID不能为空|租户ID长度不能超过255个字符" dc:"租户ID" eg:"tenant-a"`
	DeviceId string `json:"deviceId" v:"required|length:1,255#设备国标ID不能为空|设备国标ID长度不能超过255个字符" dc:"设备国标ID" eg:"34020000001320000001"`
}

// DeleteTenantDeviceBindingRes defines the response for deleting one tenant-device strategy binding.
type DeleteTenantDeviceBindingRes struct {
	TenantId string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
	DeviceId string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000001"`
}
