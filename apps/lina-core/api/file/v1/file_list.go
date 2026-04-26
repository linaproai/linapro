package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying the file list.
type ListReq struct {
	g.Meta         `path:"/file" method:"get" tags:"File Management" summary:"Get file list" dc:"Query the paginated file list, support filtering by file name, original name, suffix, upload time range, support sorting by file size and upload time" permission:"system:file:query"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	Name           string `json:"name" dc:"Filter by stored file name (fuzzy match)" eg:"20260319"`
	Original       string `json:"original" dc:"Filter by original filename (fuzzy match)" eg:"avatar"`
	Suffix         string `json:"suffix" dc:"Filter precisely by file suffix" eg:"png"`
	BeginTime      string `json:"beginTime" dc:"Upload time range starts" eg:"2026-01-01"`
	EndTime        string `json:"endTime" dc:"Upload time range ends" eg:"2026-12-31"`
	Scene          string `json:"scene" dc:"Filter by usage scenario: avatar=user avatar notice_image=notice announcement image notice_attachment=notice announcement attachment other=other, if not uploaded, query all" eg:"avatar"`
	OrderBy        string `json:"orderBy" dc:"Sorting fields: size,createdAt" eg:"createdAt"`
	OrderDirection string `json:"orderDirection" dc:"Sorting direction: asc=ascending order desc=descending order" eg:"desc"`
}

// ListRes File list response
type ListRes struct {
	List  []*ListItem `json:"list" dc:"file list" eg:"[]"`
	Total int         `json:"total" dc:"Total number of items" eg:"20"`
}

// ListItem represents a single file list item.
type ListItem struct {
	*entity.SysFile
	CreatedByName string `json:"createdByName" dc:"Uploader username" eg:"admin"`
}
