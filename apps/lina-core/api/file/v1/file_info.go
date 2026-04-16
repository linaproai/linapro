package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// InfoByIdsReq defines the request for querying file info by IDs.
type InfoByIdsReq struct {
	g.Meta `path:"/file/info/{ids}" method:"get" tags:"文件管理" summary:"根据ID查询文件信息" dc:"根据文件ID查询文件详细信息，支持批量查询（逗号分隔多个ID），用于文件回显" permission:"system:file:query"`
	Ids    string `json:"ids" v:"required" dc:"文件ID，多个用逗号分隔" eg:"1,2,3"`
}

// InfoByIdsRes 文件信息响应
type InfoByIdsRes struct {
	List []*entity.SysFile `json:"list" dc:"文件信息列表" eg:"[]"`
}
