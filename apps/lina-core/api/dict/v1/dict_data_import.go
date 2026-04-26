package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictData Import API

// DataImportReq defines the request for importing dictionary data.
type DataImportReq struct {
	g.Meta `path:"/dict/data/import" method:"post" mime:"multipart/form-data" tags:"Dictionary Management" summary:"Import dictionary data" dc:"To import dictionary data in batches through Excel files, you need to use the import template provided by the system." permission:"system:dict:add"`
}

// DataImportRes is the response structure for dictionary data import.
type DataImportRes struct {
	Success  int                  `json:"success" dc:"Number of successes" eg:"10"`
	Fail     int                  `json:"fail" dc:"Number of failed entries" eg:"2"`
	FailList []DataImportFailItem `json:"failList" dc:"Failure details" eg:"[]"`
}

// DataImportFailItem represents a failed import record.
type DataImportFailItem struct {
	Row    int    `json:"row" dc:"Line number" eg:"3"`
	Reason string `json:"reason" dc:"Reason for failure" eg:"Dictionary type does not exist"`
}

// DataImportTemplateReq defines the request for downloading import template.
type DataImportTemplateReq struct {
	g.Meta `path:"/dict/data/import-template" method:"get" tags:"Dictionary Management" summary:"Download dictionary data import template" dc:"Download the dictionary data import Excel template file, including required fields and data format instructions" permission:"system:dict:add"`
}

// DataImportTemplateRes is the response for template download.
type DataImportTemplateRes struct{}
