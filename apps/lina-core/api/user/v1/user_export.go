package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ExportReq defines the request for exporting users.
type ExportReq struct {
	g.Meta `path:"/user/export" method:"get" tags:"用户管理" summary:"导出用户数据" operLog:"export" dc:"导出用户数据为Excel文件，可指定导出特定用户或全部用户" permission:"system:user:export"`
	Ids    []int `json:"ids" dc:"导出指定用户ID列表，为空则导出全部" eg:"[1,2,3]"`
}

// ExportRes defines the response for exporting users.
type ExportRes struct{}
