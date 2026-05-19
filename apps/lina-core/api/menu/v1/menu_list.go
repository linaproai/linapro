// This file defines menu list DTOs and menu type enum values used by menu trees.

package v1

import (
	"lina-core/pkg/menutype"
	"lina-core/pkg/statusflag"

	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying the menu tree list.
type ListReq struct {
	g.Meta  `path:"/menu" method:"get" tags:"Menu Management" summary:"Get menu list" dc:"Get the menu list and return the tree structure. Supports filtering by menu name and status" permission:"system:menu:query"`
	Name    string                 `json:"name" dc:"Filter by menu name (fuzzy match)" eg:"User"`
	Status  *statusflag.Enabled    `json:"status" dc:"Filter by status: 1=normal 0=disabled" eg:"1"`
	Visible *statusflag.Visibility `json:"visible" dc:"Filter by display status: 1=show 0=hide" eg:"1"`
}

// MenuItem represents a single menu in the tree
type MenuItem struct {
	Id         int                   `json:"id" dc:"Menu ID" eg:"1"`
	ParentId   int                   `json:"parentId" dc:"Parent menu ID" eg:"0"`
	Name       string                `json:"name" dc:"Menu name (i18n supported)" eg:"System management"`
	Path       string                `json:"path" dc:"routing address" eg:"system"`
	Component  string                `json:"component" dc:"component path" eg:"system/user/index"`
	Perms      string                `json:"perms" dc:"Permission ID" eg:"system:user:list"`
	Icon       string                `json:"icon" dc:"menu icon" eg:"ant-design:setting-outlined"`
	Type       menutype.Code         `json:"type" dc:"Menu type: D=Directory M=Menu B=Button" eg:"M"`
	Sort       int                   `json:"sort" dc:"Show sort" eg:"1"`
	Visible    statusflag.Visibility `json:"visible" dc:"Whether to display: 1=show 0=hide" eg:"1"`
	Status     statusflag.Enabled    `json:"status" dc:"Status: 1=normal 0=disabled" eg:"1"`
	IsFrame    statusflag.YesNo      `json:"isFrame" dc:"Whether to external link: 1=yes 0=no" eg:"0"`
	IsCache    statusflag.YesNo      `json:"isCache" dc:"Whether to cache: 1=yes 0=no" eg:"0"`
	QueryParam string                `json:"queryParam" dc:"Route parameters (JSON format)" eg:""`
	Remark     string                `json:"remark" dc:"Remarks" eg:""`
	CreatedAt  *int64                `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1735689600000"`
	UpdatedAt  *int64                `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1735689600000"`
	Children   []*MenuItem           `json:"children" dc:"Submenu list" eg:"[]"`
}

// ListRes defines the response for querying the menu tree list.
type ListRes struct {
	List []*MenuItem `json:"list" dc:"Menu tree list" eg:"[]"`
}
