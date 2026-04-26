package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DataExportReq defines the request for exporting dictionary data.
type DataExportReq struct {
	g.Meta   `path:"/dict/data/export" method:"get" tags:"Dictionary Management" summary:"Export dictionary data" operLog:"export" dc:"Export dictionary data to Excel files, support filtering and exporting by dictionary type and label, and also support exporting records with specified IDs" permission:"system:dict:export"`
	DictType string `json:"dictType" dc:"Filter by dictionary type key" eg:"sys_user_sex"`
	Label    string `json:"label" dc:"Filter by dictionary tags (fuzzy matching)" eg:"male"`
	Ids      []int  `json:"ids" dc:"Specify the list of record IDs to be exported. If not passed, all records that meet the conditions will be exported." eg:"[1,2,3]"`
}

// DataExportRes defines the response for exporting dictionary data.
type DataExportRes struct{}
