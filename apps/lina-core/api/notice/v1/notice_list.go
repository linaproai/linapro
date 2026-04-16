package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Notice List API

// ListReq defines the request for listing notices.
type ListReq struct {
	g.Meta    `path:"/notice" method:"get" tags:"通知公告" summary:"获取通知公告列表" dc:"分页查询通知公告列表，支持按标题、类型、创建人筛选" permission:"system:notice:query"`
	PageNum   int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize  int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Title     string `json:"title" dc:"按标题筛选（模糊匹配）" eg:"系统维护"`
	Type      int    `json:"type" dc:"按类型筛选：1=通知 2=公告" eg:"1"`
	CreatedBy string `json:"createdBy" dc:"按创建人用户名筛选" eg:"admin"`
}

// ListRes Notice list response
type ListRes struct {
	List  []*ListItem `json:"list" dc:"通知公告列表" eg:"[]"`
	Total int         `json:"total" dc:"总条数" eg:"20"`
}

// ListItem Notice list item
type ListItem struct {
	*entity.SysNotice
	CreatedByName string `json:"createdByName" dc:"创建者用户名" eg:"admin"`
}
