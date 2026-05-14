// This file declares media tenant stream config list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListTenantStreamConfigsReq defines the request for querying tenant stream configs.
type ListTenantStreamConfigsReq struct {
	g.Meta   `path:"/media/tenant-stream-configs" method:"get" tags:"媒体管理" summary:"查询租户流配置列表" dc:"分页查询租户流配置，支持按租户ID或节点编号模糊筛选。" permission:"media:management:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Keyword  string `json:"keyword" dc:"按租户ID或节点编号模糊筛选" eg:"tenant-a"`
	Enable   *int   `json:"enable" dc:"启用状态：1开启，0关闭" eg:"1"`
}

// ListTenantStreamConfigsRes defines the response for querying tenant stream configs.
type ListTenantStreamConfigsRes struct {
	List  []*TenantStreamConfigListItem `json:"list" dc:"租户流配置列表" eg:"[]"`
	Total int                           `json:"total" dc:"匹配总数" eg:"1"`
}

// TenantStreamConfigListItem defines one tenant stream config row.
type TenantStreamConfigListItem struct {
	TenantId      string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
	MaxConcurrent int    `json:"maxConcurrent" dc:"最大并发数" eg:"100"`
	NodeNum       int    `json:"nodeNum" dc:"节点编号" eg:"1"`
	NodeName      string `json:"nodeName" dc:"节点名称" eg:"华东节点"`
	Enable        int    `json:"enable" dc:"1开启，0关闭" eg:"1"`
	CreatorId     int    `json:"creatorId" dc:"创建人ID" eg:"1"`
	CreateTime    string `json:"createTime" dc:"创建时间" eg:"2026-05-13 10:00:00"`
	UpdaterId     int    `json:"updaterId" dc:"修改人ID" eg:"1"`
	UpdateTime    string `json:"updateTime" dc:"修改时间" eg:"2026-05-13 10:00:00"`
}
