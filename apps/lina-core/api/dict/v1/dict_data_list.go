package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DataListReq defines the request for querying the dictionary data list.
type DataListReq struct {
	g.Meta   `path:"/dict/data" method:"get" tags:"字典管理" summary:"获取字典数据列表" dc:"分页查询字典数据列表，支持按字典类型和标签名称筛选" permission:"system:dict:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	DictType string `json:"dictType" dc:"按字典类型标识筛选" eg:"sys_user_sex"`
	Label    string `json:"label" dc:"按字典标签筛选（模糊匹配）" eg:"男"`
}

// DataListRes defines the response for querying the dictionary data list.
type DataListRes struct {
	List  []*entity.SysDictData `json:"list" dc:"字典数据列表" eg:"[]"`
	Total int                   `json:"total" dc:"总条数" eg:"3"`
}
