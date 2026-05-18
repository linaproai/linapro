package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Dict Combined Export API

// ExportReq defines the request for exporting dictionary types and data together.
type ExportReq struct {
	g.Meta `path:"/dict/export" method:"get" tags:"Dictionary Management" summary:"Export dictionary management data" operLog:"export" dc:"Export dictionary types and dictionary data to an Excel file (double Sheet format), support filtering and exporting by conditions, and also support exporting dictionary types with specified IDs" permission:"system:dict:export"`
	Name   string `json:"name" dc:"Filter by dictionary name (fuzzy matching)" eg:"gender"`
	Type   string `json:"type" dc:"Filter by dictionary type key (fuzzy matching)" eg:"sys_user"`
	Ids    []int  `json:"ids" dc:"Specify the dictionary type ID list to be exported. If not passed, all will be exported." eg:"[1,2,3]"`
}

// ExportRes is the response for dictionary combined export.
type ExportRes struct{}
