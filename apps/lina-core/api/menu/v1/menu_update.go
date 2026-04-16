package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UpdateReq defines the request for updating menu information.
type UpdateReq struct {
	g.Meta     `path:"/menu/{id}" method:"put" tags:"菜单管理" summary:"更新菜单" dc:"更新菜单信息，菜单名称在同一父级下不能与其他菜单重复" permission:"system:menu:edit"`
	Id         int    `json:"id" v:"required|min:1" dc:"菜单ID" eg:"1"`
	ParentId   *int   `json:"parentId" dc:"父菜单ID（0=根菜单）" eg:"0"`
	Name       string `json:"name" v:"required" dc:"菜单名称（支持i18n格式）" eg:"用户管理"`
	Path       string `json:"path" dc:"路由地址" eg:"user"`
	Component  string `json:"component" dc:"组件路径" eg:"system/user/index"`
	Perms      string `json:"perms" dc:"权限标识" eg:"system:user:list"`
	Icon       string `json:"icon" dc:"菜单图标" eg:"ant-design:user-outlined"`
	Type       string `json:"type" v:"required|in:D,M,B" dc:"菜单类型：D=目录 M=菜单 B=按钮" eg:"M"`
	Sort       *int   `json:"sort" dc:"显示排序" eg:"1"`
	Visible    *int   `json:"visible" v:"in:0,1" dc:"是否显示：1=显示 0=隐藏" eg:"1"`
	Status     *int   `json:"status" v:"in:0,1" dc:"状态：1=正常 0=停用" eg:"1"`
	IsFrame    *int   `json:"isFrame" v:"in:0,1" dc:"是否外链：1=是 0=否" eg:"0"`
	IsCache    *int   `json:"isCache" v:"in:0,1" dc:"是否缓存：1=是 0=否" eg:"0"`
	QueryParam string `json:"queryParam" dc:"路由参数（JSON格式）" eg:""`
	Remark     string `json:"remark" dc:"备注" eg:""`
}

// UpdateRes defines the response for updating menu information.
type UpdateRes struct{}
