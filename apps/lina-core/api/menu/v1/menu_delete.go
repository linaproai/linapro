package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeleteReq defines the request for deleting a menu.
type DeleteReq struct {
	g.Meta        `path:"/menu/{id}" method:"delete" tags:"菜单管理" summary:"删除菜单" dc:"删除菜单，如果有子菜单需要先删除子菜单或使用级联删除" permission:"system:menu:remove"`
	Id            int  `json:"id" v:"required|min:1" dc:"菜单ID" eg:"1"`
	CascadeDelete bool `json:"cascadeDelete" d:"false" dc:"是否级联删除子菜单：true=删除菜单及其所有子菜单 false=仅删除当前菜单（有子菜单时不允许删除）" eg:"false"`
}

// DeleteRes defines the response for deleting a menu.
type DeleteRes struct{}
