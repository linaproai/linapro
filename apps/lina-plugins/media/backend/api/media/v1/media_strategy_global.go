// This file declares media strategy global-state DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SetGlobalStrategyReq defines the request for setting one strategy as global.
type SetGlobalStrategyReq struct {
	g.Meta `path:"/media/strategies/{id}/global" method:"put" tags:"媒体管理" summary:"设置全局媒体策略" dc:"将指定策略设置为全局策略，并关闭其他策略的全局标记。" permission:"media:management:edit"`
	Id     int64 `json:"id" v:"required|min:1" dc:"策略ID" eg:"1"`
}

// SetGlobalStrategyRes defines the response for setting one strategy as global.
type SetGlobalStrategyRes struct {
	Id int64 `json:"id" dc:"已设置为全局的策略ID" eg:"1"`
}
