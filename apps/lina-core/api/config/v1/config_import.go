package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Config Import API

// ConfigImportReq defines the request for importing configs.
type ConfigImportReq struct {
	g.Meta `path:"/config/import" method:"post" mime:"multipart/form-data" tags:"参数设置" summary:"导入参数设置" dc:"通过Excel文件批量导入参数设置数据，需使用系统提供的导入模板" permission:"system:config:add"`
}

// ConfigImportRes is the response structure for config import.
type ConfigImportRes struct {
	Success  int                    `json:"success" dc:"成功条数" eg:"10"`
	Fail     int                    `json:"fail" dc:"失败条数" eg:"2"`
	FailList []ConfigImportFailItem `json:"failList" dc:"失败详情" eg:"[]"`
}

// ConfigImportFailItem represents a failed import record.
type ConfigImportFailItem struct {
	Row    int    `json:"row" dc:"行号" eg:"3"`
	Reason string `json:"reason" dc:"失败原因" eg:"参数键名已存在"`
}

// ConfigImportTemplateReq defines the request for downloading import template.
type ConfigImportTemplateReq struct {
	g.Meta `path:"/config/import-template" method:"get" tags:"参数设置" summary:"下载参数设置导入模板" dc:"下载参数设置导入Excel模板文件，包含必填字段和数据格式说明" permission:"system:config:add"`
}

// ConfigImportTemplateRes is the response for template download.
type ConfigImportTemplateRes struct{}
