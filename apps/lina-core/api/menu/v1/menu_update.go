package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UpdateReq defines the request for updating menu information.
type UpdateReq struct {
	g.Meta     `path:"/menu/{id}" method:"put" tags:"Menu Management" summary:"Update menu" dc:"Update the menu information. The menu name cannot be repeated with other menus under the same parent; the directory and menu icons must remain globally unique in the left navigation, and duplicate icons will be rejected." permission:"system:menu:edit"`
	Id         int    `json:"id" v:"required|min:1" dc:"Menu ID" eg:"1"`
	ParentId   *int   `json:"parentId" dc:"Parent menu ID (0=root menu)" eg:"0"`
	Name       string `json:"name" v:"required" dc:"Menu name (supports i18n format)" eg:"User Management"`
	Path       string `json:"path" dc:"routing address" eg:"user"`
	Component  string `json:"component" dc:"component path" eg:"system/user/index"`
	Perms      string `json:"perms" dc:"Permission ID" eg:"system:user:list"`
	Icon       string `json:"icon" dc:"Menu icon; when saving catalog and menu types, the left navigation icon will be verified to be globally unique. Button types ignore this constraint." eg:"ant-design:user-outlined"`
	Type       string `json:"type" v:"required|in:D,M,B" dc:"Menu type: D=Directory M=Menu B=Button" eg:"M"`
	Sort       *int   `json:"sort" dc:"Show sort" eg:"1"`
	Visible    *int   `json:"visible" v:"in:0,1" dc:"Whether to display: 1=show 0=hide" eg:"1"`
	Status     *int   `json:"status" v:"in:0,1" dc:"Status: 1=normal 0=disabled" eg:"1"`
	IsFrame    *int   `json:"isFrame" v:"in:0,1" dc:"Whether to external link: 1=yes 0=no" eg:"0"`
	IsCache    *int   `json:"isCache" v:"in:0,1" dc:"Whether to cache: 1=yes 0=no" eg:"0"`
	QueryParam string `json:"queryParam" dc:"Route parameters (JSON format)" eg:""`
	Remark     string `json:"remark" dc:"Remarks" eg:""`
}

// UpdateRes defines the response for updating menu information.
type UpdateRes struct{}
