package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TypeExportReq defines the request for exporting dictionary types.
type TypeExportReq struct {
	g.Meta `path:"/dict/type/export" method:"get" tags:"Dictionary Management" summary:"Export dictionary type" operLog:"export" dc:"Export dictionary type data to Excel files, support filtering and exporting by conditions, and also support exporting records with specified IDs" permission:"system:dict:export"`
	Name   string `json:"name" dc:"Filter by dictionary name (fuzzy matching)" eg:"gender"`
	Type   string `json:"type" dc:"Filter by dictionary type key (fuzzy matching)" eg:"sys_user"`
	Ids    []int  `json:"ids" dc:"Specify the list of record IDs to be exported. If not passed, all records that meet the conditions will be exported." eg:"[1,2,3]"`
}

// TypeExportRes defines the response for exporting dictionary types.
type TypeExportRes struct{}
