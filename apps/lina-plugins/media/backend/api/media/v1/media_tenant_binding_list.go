// This file declares media tenant strategy binding list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListTenantBindingsReq defines the request for querying tenant strategy bindings.
type ListTenantBindingsReq struct {
	g.Meta   `path:"/media/tenant-bindings" method:"get" tags:"媒体管理" summary:"查询租户策略绑定列表" dc:"分页查询租户ID与媒体策略的绑定关系。" permission:"media:management:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Keyword  string `json:"keyword" dc:"按租户ID模糊筛选" eg:"tenant-a"`
}

// ListTenantBindingsRes defines the response for querying tenant strategy bindings.
type ListTenantBindingsRes struct {
	List  []*TenantBindingItem `json:"list" dc:"租户策略绑定列表" eg:"[]"`
	Total int                  `json:"total" dc:"匹配总数" eg:"1"`
}

// TenantBindingItem defines one tenant strategy binding row.
type TenantBindingItem struct {
	RowKey       string `json:"rowKey" dc:"前端表格行唯一键" eg:"tenant:tenant-a"`
	TenantId     string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
	StrategyId   int64  `json:"strategyId" dc:"策略ID" eg:"1"`
	StrategyName string `json:"strategyName" dc:"策略名称" eg:"默认直播策略"`
}
