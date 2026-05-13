// This file declares media strategy list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListStrategiesReq defines the request for querying media strategies.
type ListStrategiesReq struct {
	g.Meta   `path:"/media/strategies" method:"get" tags:"媒体管理" summary:"查询媒体策略列表" dc:"分页查询媒体策略，支持按策略名称模糊筛选。" permission:"media:management:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Keyword  string `json:"keyword" dc:"按策略名称模糊筛选" eg:"直播"`
	Enable   int    `json:"enable" dc:"启用状态：1开启，2关闭；不传时查询全部" eg:"1"`
	Global   int    `json:"global" dc:"全局状态：1是，2否；不传时查询全部" eg:"1"`
}

// ListStrategiesRes defines the response for querying media strategies.
type ListStrategiesRes struct {
	List  []*StrategyListItem `json:"list" dc:"媒体策略列表" eg:"[]"`
	Total int                 `json:"total" dc:"匹配总数" eg:"1"`
}

// StrategyListItem defines one media strategy list row.
type StrategyListItem struct {
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
