// This file declares media strategy delete DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteStrategyReq defines the request for deleting one media strategy.
type DeleteStrategyReq struct {
	g.Meta `path:"/media/strategies/{id}" method:"delete" tags:"媒体管理" summary:"删除媒体策略" dc:"删除未被绑定引用的媒体策略。" permission:"media:management:remove"`
	Id     int64 `json:"id" v:"required|min:1" dc:"策略ID" eg:"1"`
}

// DeleteStrategyRes defines the response for deleting one media strategy.
type DeleteStrategyRes struct {
	Id int64 `json:"id" dc:"已删除策略ID" eg:"1"`
}
