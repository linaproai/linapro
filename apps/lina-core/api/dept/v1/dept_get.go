package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// GetReq defines the request for querying department detail.
type GetReq struct {
	g.Meta `path:"/dept/{id}" method:"get" tags:"部门管理" summary:"获取部门详情" dc:"根据部门ID获取部门的完整详细信息，包括基本信息、负责人、联系方式等" permission:"system:dept:query"`
	Id     int `json:"id" v:"required" dc:"部门ID" eg:"100"`
}

// GetRes is the response for department detail.
type GetRes struct {
	*entity.SysDept `dc:"部门详细信息，包含部门的所有字段数据" eg:""`
}
