// This file defines dictionary data import DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

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
