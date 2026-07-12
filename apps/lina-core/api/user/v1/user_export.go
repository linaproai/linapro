package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ExportReq defines the request for exporting users.
type ExportReq struct {
	g.Meta `path:"/user/export" method:"get" tags:"User Management" summary:"Export user data" operLog:"export" dc:"Export user data to Excel file, you can specify to export specific users or all users" permission:"system:user:export"`
	Ids    []int `json:"ids" dc:"Optional user ID list as a query array (ids[]=1&ids[]=2). When empty, export all accessible users." eg:"[1,2,3]"`
}

// ExportRes defines the response for exporting users.
type ExportRes struct{}
