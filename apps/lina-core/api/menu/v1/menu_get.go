package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GetReq defines the request for querying menu detail.
type GetReq struct {
	g.Meta `path:"/menu/{id}" method:"get" tags:"菜单管理" summary:"获取菜单详情" dc:"根据菜单ID获取菜单详情信息，包含父菜单名称" permission:"system:menu:query"`
	Id     int `json:"id" v:"required|min:1" dc:"菜单ID" eg:"1"`
}

// GetRes defines the response for querying menu detail.
type GetRes struct {
	*MenuItem  `dc:"菜单详情" eg:""`
	ParentName string `json:"parentName" dc:"父菜单名称" eg:"系统管理"`
}
