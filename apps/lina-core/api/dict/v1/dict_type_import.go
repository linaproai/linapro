package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictType Import API

// TypeImportReq defines the request for importing dictionary types.
type TypeImportReq struct {
	g.Meta `path:"/dict/type/import" method:"post" mime:"multipart/form-data" tags:"字典管理" summary:"导入字典类型" dc:"通过Excel文件批量导入字典类型数据，需使用系统提供的导入模板" permission:"system:dict:add"`
}

// TypeImportRes is the response structure for dictionary type import.
type TypeImportRes struct {
	Success  int                  `json:"success" dc:"成功条数" eg:"10"`
	Fail     int                  `json:"fail" dc:"失败条数" eg:"2"`
	FailList []TypeImportFailItem `json:"failList" dc:"失败详情" eg:"[]"`
}

// TypeImportFailItem represents a failed import record.
type TypeImportFailItem struct {
	Row    int    `json:"row" dc:"行号" eg:"3"`
	Reason string `json:"reason" dc:"失败原因" eg:"字典类型已存在"`
}

// TypeImportTemplateReq defines the request for downloading import template.
type TypeImportTemplateReq struct {
	g.Meta `path:"/dict/type/import-template" method:"get" tags:"字典管理" summary:"下载字典类型导入模板" dc:"下载字典类型导入Excel模板文件，包含必填字段和数据格式说明" permission:"system:dict:add"`
}

// TypeImportTemplateRes is the response for template download.
type TypeImportTemplateRes struct{}
