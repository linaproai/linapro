// This file declares media strategy detail DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetStrategyReq defines the request for querying one media strategy.
type GetStrategyReq struct {
	g.Meta `path:"/media/strategies/{id}" method:"get" tags:"媒体管理" summary:"获取媒体策略详情" dc:"根据策略ID获取媒体策略详情。" permission:"media:management:query"`
	Id     int64 `json:"id" v:"required|min:1" dc:"策略ID" eg:"1"`
}

// GetStrategyRes defines the media strategy detail response.
type GetStrategyRes struct {
	Id        int64  `json:"id" dc:"策略ID" eg:"1"`
	Name      string `json:"name" dc:"策略名称" eg:"默认直播策略"`
	Strategy  string `json:"strategy" dc:"YAML格式策略内容" eg:"record: true"`
	Global    int    `json:"global" dc:"是否全局策略：1是，2否" eg:"2"`
	Enable    int    `json:"enable" dc:"启用状态：1开启，2关闭" eg:"1"`
	CreatorId int64  `json:"creatorId" dc:"创建人ID" eg:"1"`
	UpdaterId int64  `json:"updaterId" dc:"修改人ID" eg:"1"`
	CreateTime string `json:"createTime" dc:"创建时间" eg:"2026-05-13 10:00:00"`
	UpdateTime string `json:"updateTime" dc:"修改时间" eg:"2026-05-13 10:00:00"`
}
