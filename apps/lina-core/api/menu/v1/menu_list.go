package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying the menu tree list.
type ListReq struct {
	g.Meta  `path:"/menu" method:"get" tags:"菜单管理" summary:"获取菜单列表" dc:"获取菜单列表，返回树形结构。支持按菜单名称、状态进行筛选" permission:"system:menu:query"`
	Name    string `json:"name" dc:"按菜单名称筛选（模糊匹配）" eg:"用户"`
	Status  *int   `json:"status" dc:"按状态筛选：1=正常 0=停用" eg:"1"`
	Visible *int   `json:"visible" dc:"按显示状态筛选：1=显示 0=隐藏" eg:"1"`
}

// MenuItem represents a single menu in the tree
type MenuItem struct {
	Id         int         `json:"id" dc:"菜单ID" eg:"1"`
	ParentId   int         `json:"parentId" dc:"父菜单ID" eg:"0"`
	Name       string      `json:"name" dc:"菜单名称（支持i18n）" eg:"系统管理"`
	Path       string      `json:"path" dc:"路由地址" eg:"system"`
	Component  string      `json:"component" dc:"组件路径" eg:"system/user/index"`
	Perms      string      `json:"perms" dc:"权限标识" eg:"system:user:list"`
	Icon       string      `json:"icon" dc:"菜单图标" eg:"ant-design:setting-outlined"`
	Type       string      `json:"type" dc:"菜单类型：D=目录 M=菜单 B=按钮" eg:"M"`
	Sort       int         `json:"sort" dc:"显示排序" eg:"1"`
	Visible    int         `json:"visible" dc:"是否显示：1=显示 0=隐藏" eg:"1"`
	Status     int         `json:"status" dc:"状态：1=正常 0=停用" eg:"1"`
	IsFrame    int         `json:"isFrame" dc:"是否外链：1=是 0=否" eg:"0"`
	IsCache    int         `json:"isCache" dc:"是否缓存：1=是 0=否" eg:"0"`
	QueryParam string      `json:"queryParam" dc:"路由参数（JSON格式）" eg:""`
	Remark     string      `json:"remark" dc:"备注" eg:""`
	CreatedAt  string      `json:"createdAt" dc:"创建时间" eg:"2025-01-01 00:00:00"`
	UpdatedAt  string      `json:"updatedAt" dc:"更新时间" eg:"2025-01-01 00:00:00"`
	Children   []*MenuItem `json:"children" dc:"子菜单列表" eg:"[]"`
}

// ListRes defines the response for querying the menu tree list.
type ListRes struct {
	List []*MenuItem `json:"list" dc:"菜单树形列表" eg:"[]"`
}
