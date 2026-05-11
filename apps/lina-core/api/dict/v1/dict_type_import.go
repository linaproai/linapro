// This file defines dictionary type import DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

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
