package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying the file list.
type ListReq struct {
	g.Meta         `path:"/file" method:"get" tags:"文件管理" summary:"获取文件列表" dc:"分页查询文件列表，支持按文件名、原始名、后缀、上传时间范围筛选，支持按文件大小和上传时间排序" permission:"system:file:query"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Name           string `json:"name" dc:"按存储文件名筛选（模糊匹配）" eg:"20260319"`
	Original       string `json:"original" dc:"按原始文件名筛选（模糊匹配）" eg:"avatar"`
	Suffix         string `json:"suffix" dc:"按文件后缀精确筛选" eg:"png"`
	BeginTime      string `json:"beginTime" dc:"上传时间范围开始" eg:"2026-01-01"`
	EndTime        string `json:"endTime" dc:"上传时间范围结束" eg:"2026-12-31"`
	Scene          string `json:"scene" dc:"按使用场景筛选：avatar=用户头像 notice_image=通知公告图片 notice_attachment=通知公告附件 other=其他，不传则查询全部" eg:"avatar"`
	OrderBy        string `json:"orderBy" dc:"排序字段：size,createdAt" eg:"createdAt"`
	OrderDirection string `json:"orderDirection" dc:"排序方向：asc=升序 desc=降序" eg:"desc"`
}

// ListRes File list response
type ListRes struct {
	List  []*ListItem `json:"list" dc:"文件列表" eg:"[]"`
	Total int         `json:"total" dc:"总条数" eg:"20"`
}

// ListItem represents a single file list item.
type ListItem struct {
	*entity.SysFile
	CreatedByName string `json:"createdByName" dc:"上传者用户名" eg:"admin"`
}
