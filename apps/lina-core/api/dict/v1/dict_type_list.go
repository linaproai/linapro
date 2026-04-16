package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// TypeListReq defines the request for querying the dictionary type list.
type TypeListReq struct {
	g.Meta   `path:"/dict/type" method:"get" tags:"字典管理" summary:"获取字典类型列表" dc:"分页查询字典类型列表，支持按字典名称和字典类型标识筛选" permission:"system:dict:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Name     string `json:"name" dc:"按字典名称筛选（模糊匹配）" eg:"用户性别"`
	Type     string `json:"type" dc:"按字典类型标识筛选（模糊匹配）" eg:"sys_user_sex"`
}

// TypeListRes defines the response for querying the dictionary type list.
type TypeListRes struct {
	List  []*entity.SysDictType `json:"list" dc:"字典类型列表" eg:"[]"`
	Total int                   `json:"total" dc:"总条数" eg:"10"`
}
