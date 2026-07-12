// This file defines menu update API DTOs and keeps optional integer flags bound
// to shared public flag contracts.

package v1

import (
	"lina-core/pkg/menutype"
	"lina-core/pkg/statusflag"

	"github.com/gogf/gf/v2/frame/g"
)

// UpdateReq defines the request for updating menu information.
// Name and Type are optional so list-row switches can submit only status or
// visible without rewriting localized display titles.
type UpdateReq struct {
	g.Meta     `path:"/menu/{id}" method:"put" tags:"Menu Management" summary:"Update menu" dc:"Update the menu information. Omitted fields keep their current values. Writing status or visible cascades that field value to all descendant menus. The menu name cannot be repeated with other menus under the same parent; the directory and menu icons must remain globally unique in the left navigation, and duplicate icons will be rejected." permission:"system:menu:edit"`
	Id         int                    `json:"id" v:"required|min:1" dc:"Menu ID" eg:"1"`
	ParentId   *int                   `json:"parentId" dc:"Parent menu ID (0=root menu)" eg:"0"`
	Name       string                 `json:"name" dc:"Menu name (supports i18n format). Optional for partial updates such as status or visibility switches." eg:"User Management"`
	Path       string                 `json:"path" dc:"routing address" eg:"user"`
	Component  string                 `json:"component" dc:"component path" eg:"system/user/index"`
	Perms      string                 `json:"perms" dc:"Permission ID" eg:"system:user:list"`
	Icon       string                 `json:"icon" dc:"Menu icon; when saving catalog and menu types, the left navigation icon will be verified to be globally unique. Button types ignore this constraint." eg:"ant-design:user-outlined"`
	Type       menutype.Code          `json:"type" v:"in:D,M,B" dc:"Menu type: D=Directory M=Menu B=Button. Optional for partial updates." eg:"M"`
	Sort       *int                   `json:"sort" dc:"Show sort" eg:"1"`
	Visible    *statusflag.Visibility `json:"visible" v:"in:0,1" dc:"Whether to display: 1=show 0=hide. Writing this field cascades the value to all descendants." eg:"1"`
	Status     *statusflag.Enabled    `json:"status" v:"in:0,1" dc:"Status: 1=normal 0=disabled. Writing this field cascades the value to all descendants." eg:"1"`
	IsFrame    *statusflag.YesNo      `json:"isFrame" v:"in:0,1" dc:"Whether to external link: 1=yes 0=no" eg:"0"`
	IsCache    *statusflag.YesNo      `json:"isCache" v:"in:0,1" dc:"Whether to cache: 1=yes 0=no" eg:"0"`
	QueryParam string                 `json:"queryParam" dc:"Route parameters (JSON format)" eg:""`
	Remark     string                 `json:"remark" dc:"Remarks" eg:""`
}

// UpdateRes defines the response for updating menu information.
type UpdateRes struct{}
