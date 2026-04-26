package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DataListReq defines the request for querying the dictionary data list.
type DataListReq struct {
	g.Meta   `path:"/dict/data" method:"get" tags:"Dictionary Management" summary:"Get dictionary data list" dc:"Paginated query dictionary data list, supports filtering by dictionary type and label" permission:"system:dict:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	DictType string `json:"dictType" dc:"Filter by dictionary type key" eg:"sys_user_sex"`
	Label    string `json:"label" dc:"Filter by dictionary tags (fuzzy matching)" eg:"male"`
}

// DataListRes defines the response for querying the dictionary data list.
type DataListRes struct {
	List  []*entity.SysDictData `json:"list" dc:"Dictionary data list" eg:"[]"`
	Total int                   `json:"total" dc:"Total number of items" eg:"3"`
}
