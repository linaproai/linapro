package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictData Import API

// DataImportReq defines the request for importing dictionary data.
type DataImportReq struct {
	g.Meta `path:"/dict/data/import" method:"post" mime:"multipart/form-data" tags:"字典管理" summary:"导入字典数据" dc:"通过Excel文件批量导入字典数据，需使用系统提供的导入模板" permission:"system:dict:add"`
}

// DataImportRes is the response structure for dictionary data import.
type DataImportRes struct {
	Success  int                  `json:"success" dc:"成功条数" eg:"10"`
	Fail     int                  `json:"fail" dc:"失败条数" eg:"2"`
	FailList []DataImportFailItem `json:"failList" dc:"失败详情" eg:"[]"`
}

// DataImportFailItem represents a failed import record.
type DataImportFailItem struct {
	Row    int    `json:"row" dc:"行号" eg:"3"`
	Reason string `json:"reason" dc:"失败原因" eg:"字典类型不存在"`
}

// DataImportTemplateReq defines the request for downloading import template.
type DataImportTemplateReq struct {
	g.Meta `path:"/dict/data/import-template" method:"get" tags:"字典管理" summary:"下载字典数据导入模板" dc:"下载字典数据导入Excel模板文件，包含必填字段和数据格式说明" permission:"system:dict:add"`
}

// DataImportTemplateRes is the response for template download.
type DataImportTemplateRes struct{}
