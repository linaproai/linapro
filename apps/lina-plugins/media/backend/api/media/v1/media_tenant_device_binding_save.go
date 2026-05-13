// This file declares media tenant-device strategy binding save DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SaveTenantDeviceBindingReq defines the request for creating or updating one tenant-device strategy binding.
type SaveTenantDeviceBindingReq struct {
	g.Meta     `path:"/media/tenant-device-bindings/{tenantId}/{deviceId}" method:"put" tags:"媒体管理" summary:"保存租户设备策略绑定" dc:"按租户ID和设备国标ID创建或替换租户设备策略绑定。" permission:"media:management:edit"`
	TenantId   string `json:"tenantId" v:"required|length:1,255#租户ID不能为空|租户ID长度不能超过255个字符" dc:"租户ID" eg:"tenant-a"`
	DeviceId   string `json:"deviceId" v:"required|length:1,255#设备国标ID不能为空|设备国标ID长度不能超过255个字符" dc:"设备国标ID" eg:"34020000001320000001"`
	StrategyId int64  `json:"strategyId" v:"required|min:1#策略ID不能为空|策略ID必须大于0" dc:"策略ID" eg:"1"`
}

// SaveTenantDeviceBindingRes defines the response for saving one tenant-device strategy binding.
type SaveTenantDeviceBindingRes struct {
	TenantId   string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
	DeviceId   string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000001"`
	StrategyId int64  `json:"strategyId" dc:"策略ID" eg:"1"`
}
