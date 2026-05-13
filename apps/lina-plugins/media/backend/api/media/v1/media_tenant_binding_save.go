// This file declares media tenant strategy binding save DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SaveTenantBindingReq defines the request for creating or updating one tenant strategy binding.
type SaveTenantBindingReq struct {
	g.Meta     `path:"/media/tenant-bindings/{tenantId}" method:"put" tags:"媒体管理" summary:"保存租户策略绑定" dc:"按租户ID创建或替换租户策略绑定。" permission:"media:management:edit"`
	TenantId   string `json:"tenantId" v:"required|length:1,255#租户ID不能为空|租户ID长度不能超过255个字符" dc:"租户ID" eg:"tenant-a"`
	StrategyId int64  `json:"strategyId" v:"required|min:1#策略ID不能为空|策略ID必须大于0" dc:"策略ID" eg:"1"`
}

// SaveTenantBindingRes defines the response for saving one tenant strategy binding.
type SaveTenantBindingRes struct {
	TenantId   string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
	StrategyId int64  `json:"strategyId" dc:"策略ID" eg:"1"`
}
