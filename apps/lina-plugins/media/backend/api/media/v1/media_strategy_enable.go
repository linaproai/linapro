// This file declares media strategy enable-state DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateStrategyEnableReq defines the request for changing strategy enable status.
type UpdateStrategyEnableReq struct {
	g.Meta `path:"/media/strategies/{id}/enable" method:"put" tags:"媒体管理" summary:"修改媒体策略启用状态" dc:"修改指定媒体策略的启用状态。" permission:"media:management:edit"`
	Id     int64 `json:"id" v:"required|min:1" dc:"策略ID" eg:"1"`
	Enable int   `json:"enable" v:"required|in:1,2#启用状态不能为空|启用状态只能是1或2" dc:"启用状态：1开启，2关闭" eg:"1"`
}

// UpdateStrategyEnableRes defines the response for changing strategy enable status.
type UpdateStrategyEnableRes struct {
	Id int64 `json:"id" dc:"已更新策略ID" eg:"1"`
}
