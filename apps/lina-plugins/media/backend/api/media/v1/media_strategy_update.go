// This file declares media strategy update DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateStrategyReq defines the request for updating one media strategy.
type UpdateStrategyReq struct {
	g.Meta   `path:"/media/strategies/{id}" method:"put" tags:"媒体管理" summary:"修改媒体策略" dc:"修改指定媒体策略的名称、策略内容、启用状态或全局状态。" permission:"media:management:edit"`
	Id       int64  `json:"id" v:"required|min:1" dc:"策略ID" eg:"1"`
	Name     string `json:"name" v:"required|length:1,255#策略名称不能为空|策略名称长度不能超过255个字符" dc:"策略名称" eg:"默认直播策略"`
	Strategy string `json:"strategy" v:"required#策略内容不能为空" dc:"YAML格式策略内容" eg:"record: true"`
	Enable   int    `json:"enable" d:"1" dc:"启用状态：1开启，2关闭" eg:"1"`
	Global   int    `json:"global" d:"2" dc:"是否全局策略：1是，2否" eg:"2"`
}

// UpdateStrategyRes defines the response for updating one media strategy.
type UpdateStrategyRes struct {
	Id int64 `json:"id" dc:"已更新策略ID" eg:"1"`
}
