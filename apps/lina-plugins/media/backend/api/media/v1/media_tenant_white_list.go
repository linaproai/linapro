// This file declares media tenant whitelist list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListTenantWhitesReq defines the request for querying tenant whitelist entries.
type ListTenantWhitesReq struct {
	g.Meta   `path:"/media/tenant-whites" method:"get" tags:"媒体管理" summary:"查询租户白名单列表" dc:"分页查询租户白名单，支持按租户ID、白名单地址或描述模糊筛选。" permission:"media:management:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Keyword  string `json:"keyword" dc:"按租户ID、白名单地址或描述模糊筛选" eg:"tenant-a"`
	Enable   *int   `json:"enable" dc:"启用状态：1开启，0关闭" eg:"1"`
}

// ListTenantWhitesRes defines the response for querying tenant whitelist entries.
type ListTenantWhitesRes struct {
	List  []*TenantWhiteListItem `json:"list" dc:"租户白名单列表" eg:"[]"`
	Total int                    `json:"total" dc:"匹配总数" eg:"1"`
}

// TenantWhiteListItem defines one tenant whitelist row.
type TenantWhiteListItem struct {
	TenantId    string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
	Ip          string `json:"ip" dc:"白名单地址" eg:"192.168.1.10"`
	Description string `json:"description" dc:"白名单描述" eg:"总部出口"`
	Enable      int    `json:"enable" dc:"1开启，0关闭" eg:"1"`
	CreatorId   int    `json:"creatorId" dc:"创建人ID" eg:"1"`
	CreateTime  string `json:"createTime" dc:"创建时间" eg:"2026-05-13 10:00:00"`
	UpdaterId   int    `json:"updaterId" dc:"修改人ID" eg:"1"`
	UpdateTime  string `json:"updateTime" dc:"修改时间" eg:"2026-05-13 10:00:00"`
}
