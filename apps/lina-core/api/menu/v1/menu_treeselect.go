package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TreeSelectReq defines the request for querying the menu tree select data.
type TreeSelectReq struct {
	g.Meta `path:"/menu/treeselect" method:"get" tags:"Menu Management" summary:"Get menu dropdown tree" dc:"Get the menu dropdown tree, used for selection when assigning the role to the menu. Filter out button type menus" permission:"system:menu:query"`
}

// MenuTreeNode represents a node in the tree select
type MenuTreeNode struct {
	Id       int             `json:"id" dc:"Menu ID" eg:"1"`
	ParentId int             `json:"parentId" dc:"Parent menu ID" eg:"0"`
	Label    string          `json:"label" dc:"Menu name" eg:"System management"`
	Type     string          `json:"type" dc:"Menu type: D=Directory M=Menu B=Button" eg:"D"`
	Icon     string          `json:"icon" dc:"menu icon" eg:"ant-design:dashboard-outlined"`
	Children []*MenuTreeNode `json:"children" dc:"submenu" eg:"[]"`
}

// TreeSelectRes defines the response for querying the menu tree select data.
type TreeSelectRes struct {
	List []*MenuTreeNode `json:"list" dc:"Menu tree list" eg:"[]"`
}
