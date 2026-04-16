package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Dict Combined Import API

// ImportReq defines the request for importing dictionary types and data together.
type ImportReq struct {
	g.Meta `path:"/dict/import" method:"post" mime:"multipart/form-data" tags:"字典管理" summary:"导入字典管理数据" dc:"通过Excel文件批量导入字典类型和字典数据，需使用系统提供的导入模板。文件包含两个Sheet：字典类型和字典数据" permission:"system:dict:add"`
}

// ImportRes is the response structure for dictionary combined import.
type ImportRes struct {
	TypeSuccess int              `json:"typeSuccess" dc:"字典类型成功导入条数" eg:"10"`
	TypeFail    int              `json:"typeFail" dc:"字典类型失败条数" eg:"2"`
	DataSuccess int              `json:"dataSuccess" dc:"字典数据成功导入条数" eg:"50"`
	DataFail    int              `json:"dataFail" dc:"字典数据失败条数" eg:"5"`
	FailList    []ImportFailItem `json:"failList" dc:"失败详情" eg:"[]"`
}

// ImportFailItem represents a failed import record.
type ImportFailItem struct {
	Sheet  string `json:"sheet" dc:"Sheet名称（字典类型/字典数据）" eg:"字典类型"`
	Row    int    `json:"row" dc:"行号" eg:"3"`
	Reason string `json:"reason" dc:"失败原因" eg:"字典类型已存在"`
}

// ImportTemplateReq defines the request for downloading combined import template.
type ImportTemplateReq struct {
	g.Meta `path:"/dict/import-template" method:"get" tags:"字典管理" summary:"下载字典管理导入模板" dc:"下载字典管理导入Excel模板文件，包含字典类型和字典数据两个Sheet，每个Sheet包含示例数据和字段说明" permission:"system:dict:add"`
}

// ImportTemplateRes is the response for template download.
type ImportTemplateRes struct{}
