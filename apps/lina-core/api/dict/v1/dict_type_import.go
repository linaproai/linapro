package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictType Import API

// TypeImportReq defines the request for importing dictionary types.
type TypeImportReq struct {
	g.Meta `path:"/dict/type/import" method:"post" mime:"multipart/form-data" tags:"Dictionary Management" summary:"Import dictionary type" dc:"To batch import dictionary type data through Excel files, you need to use the import template provided by the system." permission:"system:dict:add"`
}

// TypeImportRes is the response structure for dictionary type import.
type TypeImportRes struct {
	Success  int                  `json:"success" dc:"Number of successes" eg:"10"`
	Fail     int                  `json:"fail" dc:"Number of failed entries" eg:"2"`
	FailList []TypeImportFailItem `json:"failList" dc:"Failure details" eg:"[]"`
}

// TypeImportFailItem represents a failed import record.
type TypeImportFailItem struct {
	Row    int    `json:"row" dc:"Line number" eg:"3"`
	Reason string `json:"reason" dc:"Reason for failure" eg:"Dictionary type already exists"`
}

// TypeImportTemplateReq defines the request for downloading import template.
type TypeImportTemplateReq struct {
	g.Meta `path:"/dict/type/import-template" method:"get" tags:"Dictionary Management" summary:"Download dictionary type import template" dc:"Download the dictionary type import Excel template file, including required fields and data format instructions" permission:"system:dict:add"`
}

// TypeImportTemplateRes is the response for template download.
type TypeImportTemplateRes struct{}
