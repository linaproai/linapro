// This file declares media device strategy binding list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListDeviceBindingsReq defines the request for querying device strategy bindings.
type ListDeviceBindingsReq struct {
	g.Meta   `path:"/media/device-bindings" method:"get" tags:"媒体管理" summary:"查询设备策略绑定列表" dc:"分页查询设备国标ID与媒体策略的绑定关系。" permission:"media:management:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Keyword  string `json:"keyword" dc:"按设备国标ID模糊筛选" eg:"34020000001320000001"`
}

// ListDeviceBindingsRes defines the response for querying device strategy bindings.
type ListDeviceBindingsRes struct {
	List  []*DeviceBindingItem `json:"list" dc:"设备策略绑定列表" eg:"[]"`
	Total int                  `json:"total" dc:"匹配总数" eg:"1"`
}

// DeviceBindingItem defines one device strategy binding row.
type DeviceBindingItem struct {
	RowKey       string `json:"rowKey" dc:"前端表格行唯一键" eg:"device:34020000001320000001"`
	DeviceId     string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000001"`
	StrategyId   int64  `json:"strategyId" dc:"策略ID" eg:"1"`
	StrategyName string `json:"strategyName" dc:"策略名称" eg:"默认直播策略"`
}
