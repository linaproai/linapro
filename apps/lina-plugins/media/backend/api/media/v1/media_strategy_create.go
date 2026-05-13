// This file declares media strategy create DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateStrategyReq defines the request for creating one media strategy.
type CreateStrategyReq struct {
	g.Meta   `path:"/media/strategies" method:"post" tags:"媒体管理" summary:"新增媒体策略" dc:"新增一条媒体策略，策略内容按YAML文本保存。" permission:"media:management:add"`
	Name     string `json:"name" v:"required|length:1,255#策略名称不能为空|策略名称长度不能超过255个字符" dc:"策略名称" eg:"默认直播策略"`
	Strategy string `json:"strategy" v:"required#策略内容不能为空" dc:"YAML格式策略内容" eg:"record: true"`
	Enable   int    `json:"enable" d:"1" dc:"启用状态：1开启，2关闭" eg:"1"`
	Global   int    `json:"global" d:"2" dc:"是否全局策略：1是，2否" eg:"2"`
}

// CreateStrategyRes defines the response for creating one media strategy.
type CreateStrategyRes struct {
	Id int64 `json:"id" dc:"新建策略ID" eg:"1"`
}
