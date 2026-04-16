package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UsageScenesReq defines the request for querying file usage scenes.
type UsageScenesReq struct {
	g.Meta `path:"/file/scenes" method:"get" tags:"文件管理" summary:"获取文件使用场景列表" dc:"查询所有已使用的文件使用场景标识列表，供管理工作台的筛选器、治理视图或文件表单复用" permission:"system:file:query"`
}

// UsageScenesRes File usage scenes list response
type UsageScenesRes struct {
	List []*UsageSceneItem `json:"list" dc:"使用场景列表" eg:"[]"`
}

// UsageSceneItem File usage scene item
type UsageSceneItem struct {
	Value string `json:"value" dc:"使用场景标识" eg:"avatar"`
	Label string `json:"label" dc:"使用场景名称" eg:"用户头像"`
}
