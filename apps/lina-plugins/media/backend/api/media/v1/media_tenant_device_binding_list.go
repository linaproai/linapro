// This file declares media tenant-device strategy binding list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListTenantDeviceBindingsReq defines the request for querying tenant-device strategy bindings.
type ListTenantDeviceBindingsReq struct {
	g.Meta   `path:"/media/tenant-device-bindings" method:"get" tags:"媒体管理" summary:"查询租户设备策略绑定列表" dc:"分页查询租户下设备与媒体策略的绑定关系。" permission:"media:management:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Keyword  string `json:"keyword" dc:"按租户ID或设备国标ID模糊筛选" eg:"tenant-a"`
}

// ListTenantDeviceBindingsRes defines the response for querying tenant-device strategy bindings.
type ListTenantDeviceBindingsRes struct {
	List  []*TenantDeviceBindingItem `json:"list" dc:"租户设备策略绑定列表" eg:"[]"`
	Total int                        `json:"total" dc:"匹配总数" eg:"1"`
}

// TenantDeviceBindingItem defines one tenant-device strategy binding row.
type TenantDeviceBindingItem struct {
	RowKey       string `json:"rowKey" dc:"前端表格行唯一键" eg:"tenantDevice:tenant-a:34020000001320000001"`
	TenantId     string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
	DeviceId     string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000001"`
	StrategyId   int64  `json:"strategyId" dc:"策略ID" eg:"1"`
	StrategyName string `json:"strategyName" dc:"策略名称" eg:"默认直播策略"`
}
