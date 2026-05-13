// This file declares media strategy resolution DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResolveStrategyReq defines the request for resolving one effective strategy.
type ResolveStrategyReq struct {
	g.Meta   `path:"/media/strategies/resolve" method:"get" tags:"媒体管理" summary:"解析媒体策略" dc:"按租户设备、设备、租户、全局优先级解析设备当前生效策略。" permission:"media:management:query"`
	TenantId string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
	DeviceId string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000001"`
}

// ResolveStrategyRes defines the response for resolving one effective strategy.
type ResolveStrategyRes struct {
	Matched      bool   `json:"matched" dc:"是否匹配到策略" eg:"true"`
	Source       string `json:"source" dc:"策略来源：tenantDevice、device、tenant、global或none" eg:"device"`
	SourceLabel  string `json:"sourceLabel" dc:"策略来源中文说明" eg:"设备策略"`
	StrategyId   int64  `json:"strategyId" dc:"策略ID" eg:"1"`
	StrategyName string `json:"strategyName" dc:"策略名称" eg:"默认直播策略"`
	Strategy     string `json:"strategy" dc:"YAML格式策略内容" eg:"record: true"`
}
