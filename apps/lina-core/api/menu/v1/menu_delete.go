package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeleteReq defines the request for deleting a menu.
type DeleteReq struct {
	g.Meta        `path:"/menu/{id}" method:"delete" tags:"Menu Management" summary:"delete menu" dc:"Delete the menu. If there is a submenu, you need to delete the submenu first or use cascade deletion." permission:"system:menu:remove"`
	Id            int  `json:"id" v:"required|min:1" dc:"Menu ID" eg:"1"`
	CascadeDelete bool `json:"cascadeDelete" d:"false" dc:"Whether to cascade delete submenus: true=delete the menu and all its submenus false=only delete the current menu (deletion is not allowed when there are submenus)" eg:"false"`
}

// DeleteRes defines the response for deleting a menu.
type DeleteRes struct{}
