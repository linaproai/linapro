package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CreateReq defines the request for creating a menu.
type CreateReq struct {
	g.Meta     `path:"/menu" method:"post" tags:"菜单管理" summary:"创建菜单" dc:"创建新菜单，支持目录、菜单、按钮三种类型。菜单名称在同一父级下不能重复" permission:"system:menu:add"`
	ParentId   int    `json:"parentId" d:"0" dc:"父菜单ID（0=根菜单）" eg:"0"`
	Name       string `json:"name" v:"required" dc:"菜单名称（支持i18n格式如 menu.system.user）" eg:"用户管理"`
	Path       string `json:"path" dc:"路由地址（目录和菜单类型必填）" eg:"user"`
	Component  string `json:"component" dc:"组件路径（菜单类型必填）" eg:"system/user/index"`
	Perms      string `json:"perms" dc:"权限标识（菜单和按钮类型必填）" eg:"system:user:list"`
	Icon       string `json:"icon" dc:"菜单图标" eg:"ant-design:user-outlined"`
	Type       string `json:"type" v:"required|in:D,M,B" dc:"菜单类型：D=目录 M=菜单 B=按钮" eg:"M"`
	Sort       int    `json:"sort" d:"0" dc:"显示排序（数字越小越靠前）" eg:"1"`
	Visible    int    `json:"visible" d:"1" v:"in:0,1" dc:"是否显示：1=显示 0=隐藏" eg:"1"`
	Status     int    `json:"status" d:"1" v:"in:0,1" dc:"状态：1=正常 0=停用" eg:"1"`
	IsFrame    int    `json:"isFrame" d:"0" v:"in:0,1" dc:"是否外链：1=是 0=否" eg:"0"`
	IsCache    int    `json:"isCache" d:"0" v:"in:0,1" dc:"是否缓存：1=是 0=否" eg:"0"`
	QueryParam string `json:"queryParam" dc:"路由参数（JSON格式）" eg:"{\"key\":\"value\"}"`
	Remark     string `json:"remark" dc:"备注" eg:""`
}

// CreateRes defines the response for creating a menu.
type CreateRes struct {
	Id int `json:"id" dc:"创建的菜单ID" eg:"1"`
}
