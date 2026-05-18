package v1

import "github.com/gogf/gf/v2/frame/g"

// TypeListReq defines the request for querying the dictionary type list.
type TypeListReq struct {
	g.Meta   `path:"/dict/type" method:"get" tags:"Dictionary Management" summary:"Get a list of dictionary types" dc:"Paginated query dictionary type list, supports filtering by dictionary name and dictionary type key" permission:"system:dict:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	Name     string `json:"name" dc:"Filter by dictionary name (fuzzy matching)" eg:"User gender"`
	Type     string `json:"type" dc:"Filter by dictionary type key (fuzzy matching)" eg:"sys_user_sex"`
}

// TypeListRes defines the response for querying the dictionary type list.
type TypeListRes struct {
	List  []*DictTypeItem `json:"list" dc:"List of dictionary types" eg:"[]"`
	Total int             `json:"total" dc:"Total number of items" eg:"10"`
}
