package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TypeExportReq defines the request for exporting dictionary types.
type TypeExportReq struct {
	g.Meta `path:"/dict/type/export" method:"get" tags:"字典管理" summary:"导出字典类型" operLog:"export" dc:"导出字典类型数据为Excel文件，支持按条件筛选导出，也支持导出指定ID的记录" permission:"system:dict:export"`
	Name   string `json:"name" dc:"按字典名称筛选（模糊匹配）" eg:"性别"`
	Type   string `json:"type" dc:"按字典类型标识筛选（模糊匹配）" eg:"sys_user"`
	Ids    []int  `json:"ids" dc:"指定导出的记录ID列表，不传则导出全部符合条件的记录" eg:"[1,2,3]"`
}

// TypeExportRes defines the response for exporting dictionary types.
type TypeExportRes struct{}
