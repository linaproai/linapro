package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Config Import API

// ConfigImportReq defines the request for importing configs.
type ConfigImportReq struct {
	g.Meta `path:"/config/import" method:"post" mime:"multipart/form-data" tags:"Parameter Settings" summary:"Import parameter settings" dc:"To import parameter setting data in batches through Excel files, you need to use the import template provided by the system." permission:"system:config:add"`
}

// ConfigImportRes is the response structure for config import.
type ConfigImportRes struct {
	Success  int                    `json:"success" dc:"Number of successes" eg:"10"`
	Fail     int                    `json:"fail" dc:"Number of failed entries" eg:"2"`
	FailList []ConfigImportFailItem `json:"failList" dc:"Failure details" eg:"[]"`
}

// ConfigImportFailItem represents a failed import record.
type ConfigImportFailItem struct {
	Row    int    `json:"row" dc:"Line number" eg:"3"`
	Reason string `json:"reason" dc:"Reason for failure" eg:"Parameter key name already exists"`
}

// ConfigImportTemplateReq defines the request for downloading import template.
type ConfigImportTemplateReq struct {
	g.Meta `path:"/config/import-template" method:"get" tags:"Parameter Settings" summary:"Download parameter setting import template" dc:"Download the parameter settings and import the Excel template file, including required fields and data format instructions" permission:"system:config:add"`
}

// ConfigImportTemplateRes is the response for template download.
type ConfigImportTemplateRes struct{}
