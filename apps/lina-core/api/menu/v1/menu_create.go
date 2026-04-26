package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CreateReq defines the request for creating a menu.
type CreateReq struct {
	g.Meta     `path:"/menu" method:"post" tags:"Menu Management" summary:"Create menu" dc:"Create a new menu, supporting three types: directory, menu, and button. Menu names cannot be repeated under the same parent; the directory/menu icons in the left navigation must be globally unique. Duplicate icons will be refused to save." permission:"system:menu:add"`
	ParentId   int    `json:"parentId" d:"0" dc:"Parent menu ID (0=root menu)" eg:"0"`
	Name       string `json:"name" v:"required" dc:"Menu name (supports i18n format such as menu.system.user)" eg:"User Management"`
	Path       string `json:"path" dc:"Routing address (required for directory and menu types)" eg:"user"`
	Component  string `json:"component" dc:"Component path (required for menu type)" eg:"system/user/index"`
	Perms      string `json:"perms" dc:"Permission ID (required for menu and button types)" eg:"system:user:list"`
	Icon       string `json:"icon" dc:"Menu icon; when saving catalog and menu types, the left navigation icon will be verified to be globally unique. Button types ignore this constraint." eg:"ant-design:user-outlined"`
	Type       string `json:"type" v:"required|in:D,M,B" dc:"Menu type: D=Directory M=Menu B=Button" eg:"M"`
	Sort       int    `json:"sort" d:"0" dc:"Display sorting (the smaller the number, the higher it is)" eg:"1"`
	Visible    int    `json:"visible" d:"1" v:"in:0,1" dc:"Whether to display: 1=show 0=hide" eg:"1"`
	Status     int    `json:"status" d:"1" v:"in:0,1" dc:"Status: 1=normal 0=disabled" eg:"1"`
	IsFrame    int    `json:"isFrame" d:"0" v:"in:0,1" dc:"Whether to external link: 1=yes 0=no" eg:"0"`
	IsCache    int    `json:"isCache" d:"0" v:"in:0,1" dc:"Whether to cache: 1=yes 0=no" eg:"0"`
	QueryParam string `json:"queryParam" dc:"Route parameters (JSON format)" eg:"{\"key\":\"value\"}"`
	Remark     string `json:"remark" dc:"Remarks" eg:""`
}

// CreateRes defines the response for creating a menu.
type CreateRes struct {
	Id int `json:"id" dc:"Created menu ID" eg:"1"`
}
