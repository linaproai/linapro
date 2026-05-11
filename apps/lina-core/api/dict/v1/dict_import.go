// This file defines combined dictionary import DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ImportReq defines the request for importing dictionary types and data together.
type ImportReq struct {
	g.Meta `path:"/dict/import" method:"post" mime:"multipart/form-data" tags:"Dictionary Management" summary:"Import dictionary management data" dc:"To batch import dictionary types and dictionary data through Excel files, you need to use the import template provided by the system. The file contains two Sheets: dictionary type and dictionary data" permission:"system:dict:add"`
}

// ImportRes is the response structure for dictionary combined import.
type ImportRes struct {
	TypeSuccess int              `json:"typeSuccess" dc:"Number of entries successfully imported into dictionary type" eg:"10"`
	TypeFail    int              `json:"typeFail" dc:"Number of failed entries in dictionary type" eg:"2"`
	DataSuccess int              `json:"dataSuccess" dc:"Number of dictionary data successfully imported" eg:"50"`
	DataFail    int              `json:"dataFail" dc:"Number of failed dictionary data entries" eg:"5"`
	FailList    []ImportFailItem `json:"failList" dc:"Failure details" eg:"[]"`
}

// ImportFailItem represents a failed import record.
type ImportFailItem struct {
	Sheet  string `json:"sheet" dc:"Sheet name (dictionary type/dictionary data)" eg:"dictionary type"`
	Row    int    `json:"row" dc:"Line number" eg:"3"`
	Reason string `json:"reason" dc:"Reason for failure" eg:"Dictionary type already exists"`
}
