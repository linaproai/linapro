// This file declares media device strategy binding save DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SaveDeviceBindingReq defines the request for creating or updating one device strategy binding.
type SaveDeviceBindingReq struct {
	g.Meta     `path:"/media/device-bindings/{deviceId}" method:"put" tags:"媒体管理" summary:"保存设备策略绑定" dc:"按设备国标ID创建或替换设备策略绑定。" permission:"media:management:edit"`
	DeviceId   string `json:"deviceId" v:"required|length:1,255#设备国标ID不能为空|设备国标ID长度不能超过255个字符" dc:"设备国标ID" eg:"34020000001320000001"`
	StrategyId int64  `json:"strategyId" v:"required|min:1#策略ID不能为空|策略ID必须大于0" dc:"策略ID" eg:"1"`
}

// SaveDeviceBindingRes defines the response for saving one device strategy binding.
type SaveDeviceBindingRes struct {
	DeviceId   string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000001"`
	StrategyId int64  `json:"strategyId" dc:"策略ID" eg:"1"`
}
