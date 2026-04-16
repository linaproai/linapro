package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TreeSelectReq defines the request for querying the menu tree select data.
type TreeSelectReq struct {
	g.Meta `path:"/menu/treeselect" method:"get" tags:"菜单管理" summary:"获取菜单下拉树" dc:"获取菜单下拉树，用于角色分配菜单时选择。过滤掉按钮类型的菜单" permission:"system:menu:query"`
}

// MenuTreeNode represents a node in the tree select
type MenuTreeNode struct {
	Id       int             `json:"id" dc:"菜单ID" eg:"1"`
	ParentId int             `json:"parentId" dc:"父菜单ID" eg:"0"`
	Label    string          `json:"label" dc:"菜单名称" eg:"系统管理"`
	Type     string          `json:"type" dc:"菜单类型：D=目录 M=菜单 B=按钮" eg:"D"`
	Icon     string          `json:"icon" dc:"菜单图标" eg:"ant-design:dashboard-outlined"`
	Children []*MenuTreeNode `json:"children" dc:"子菜单" eg:"[]"`
}

// TreeSelectRes defines the response for querying the menu tree select data.
type TreeSelectRes struct {
	List []*MenuTreeNode `json:"list" dc:"菜单树形列表" eg:"[]"`
}
