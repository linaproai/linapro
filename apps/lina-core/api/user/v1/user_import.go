package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ImportReq defines the request for importing users.
type ImportReq struct {
	g.Meta `path:"/user/import" method:"post" mime:"multipart/form-data" tags:"User Management" summary:"Import user data" dc:"To import user data in batches through Excel files, you need to use the import template provided by the system." permission:"system:user:import"`
}

// ImportRes is the response structure for user import.
type ImportRes struct {
	Success  int              `json:"success" dc:"Number of successes" eg:"10"`
	Fail     int              `json:"fail" dc:"Number of failed entries" eg:"2"`
	FailList []ImportFailItem `json:"failList" dc:"Failure details" eg:"[]"`
}

// ImportFailItem represents a failed import record.
type ImportFailItem struct {
	Row    int    `json:"row" dc:"Line number" eg:"3"`
	Reason string `json:"reason" dc:"Reason for failure" eg:"Username already exists"`
}

// ImportTemplateReq defines the request for downloading the user import template.
type ImportTemplateReq struct {
	g.Meta `path:"/user/import-template" method:"get" tags:"User Management" summary:"Download import template" dc:"Download the user import Excel template file, including required fields and data format instructions" permission:"system:user:import"`
}

// ImportTemplateRes defines the response for downloading the user import template.
type ImportTemplateRes struct{}
