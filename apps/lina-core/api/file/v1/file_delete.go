package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeleteReq defines the request for deleting files.
type DeleteReq struct {
	g.Meta `path:"/file/{ids}" method:"delete" tags:"文件管理" summary:"删除文件" dc:"根据文件ID删除文件，支持批量删除（逗号分隔多个ID）。同时删除物理文件和数据库记录" permission:"system:file:remove"`
	Ids    string `json:"ids" v:"required" dc:"文件ID，多个用逗号分隔" eg:"1,2,3"`
}

// DeleteRes File delete response
type DeleteRes struct{}
