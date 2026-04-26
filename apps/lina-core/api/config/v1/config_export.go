package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Config Export API

// ExportReq defines the request for exporting configs to Excel.
type ExportReq struct {
	g.Meta    `path:"/config/export" method:"get" tags:"Parameter Settings" summary:"Export parameter settings" operLog:"export" dc:"Export parameter setting data to Excel files, support filtering and exporting by conditions, and also support exporting records with specified IDs" permission:"system:config:export"`
	Name      string `json:"name" dc:"Filter by parameter name (fuzzy matching)" eg:"Main frame page"`
	Key       string `json:"key" dc:"Filter by parameter key name (fuzzy matching)" eg:"sys.index"`
	BeginTime string `json:"beginTime" dc:"Creation time range-start time, format YYYY-MM-DD" eg:"2025-01-01"`
	EndTime   string `json:"endTime" dc:"Creation time range-end time, format YYYY-MM-DD" eg:"2025-12-31"`
	Ids       []int  `json:"ids" dc:"Specify the list of record IDs to be exported. If not passed, all records that meet the conditions will be exported." eg:"[1,2,3]"`
}

// ExportRes is the config export response.
type ExportRes struct{}
