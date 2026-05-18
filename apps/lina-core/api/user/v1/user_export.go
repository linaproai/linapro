package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ExportReq defines the request for exporting users.
type ExportReq struct {
	g.Meta `path:"/user/export" method:"get" tags:"User Management" summary:"Export user data" operLog:"export" dc:"Export user data to Excel file, you can specify to export specific users or all users" permission:"system:user:export"`
	Ids    []int `json:"ids" dc:"Export the specified user ID list, if it is empty, export all" eg:"[1,2,3]"`
}

// ExportRes defines the response for exporting users.
type ExportRes struct{}
