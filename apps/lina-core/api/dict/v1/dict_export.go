package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Dict Combined Export API

// ExportReq defines the request for exporting dictionary types and data together.
type ExportReq struct {
	g.Meta `path:"/dict/export" method:"get" tags:"字典管理" summary:"导出字典管理数据" operLog:"export" dc:"导出字典类型和字典数据到一个Excel文件（双Sheet格式），支持按条件筛选导出，也支持导出指定ID的字典类型" permission:"system:dict:export"`
	Name   string `json:"name" dc:"按字典名称筛选（模糊匹配）" eg:"性别"`
	Type   string `json:"type" dc:"按字典类型标识筛选（模糊匹配）" eg:"sys_user"`
	Ids    []int  `json:"ids" dc:"指定导出的字典类型ID列表，不传则导出全部" eg:"[1,2,3]"`
}

// ExportRes is the response for dictionary combined export.
type ExportRes struct{}
