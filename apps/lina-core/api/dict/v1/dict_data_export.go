package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DataExportReq defines the request for exporting dictionary data.
type DataExportReq struct {
	g.Meta   `path:"/dict/data/export" method:"get" tags:"字典管理" summary:"导出字典数据" operLog:"export" dc:"导出字典数据为Excel文件，支持按字典类型和标签筛选导出，也支持导出指定ID的记录" permission:"system:dict:export"`
	DictType string `json:"dictType" dc:"按字典类型标识筛选" eg:"sys_user_sex"`
	Label    string `json:"label" dc:"按字典标签筛选（模糊匹配）" eg:"男"`
	Ids      []int  `json:"ids" dc:"指定导出的记录ID列表，不传则导出全部符合条件的记录" eg:"[1,2,3]"`
}

// DataExportRes defines the response for exporting dictionary data.
type DataExportRes struct{}
