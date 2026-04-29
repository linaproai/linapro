// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysMenu is the golang structure for table sys_menu.
type SysMenu struct {
	Id         int         `json:"id"         orm:"id"          description:"Menu ID"`
	ParentId   int         `json:"parentId"   orm:"parent_id"   description:"Parent menu ID, 0 means root menu"`
	MenuKey    string      `json:"menuKey"    orm:"menu_key"    description:"Stable menu business key"`
	Name       string      `json:"name"       orm:"name"        description:"Menu name with i18n support"`
	Path       string      `json:"path"       orm:"path"        description:"Route path"`
	Component  string      `json:"component"  orm:"component"   description:"Component path"`
	Perms      string      `json:"perms"      orm:"perms"       description:"Permission identifier"`
	Icon       string      `json:"icon"       orm:"icon"        description:"Menu icon"`
	Type       string      `json:"type"       orm:"type"        description:"Menu type: D=directory, M=menu, B=button"`
	Sort       int         `json:"sort"       orm:"sort"        description:"Display order"`
	Visible    int         `json:"visible"    orm:"visible"     description:"Visibility: 1=visible, 0=hidden"`
	Status     int         `json:"status"     orm:"status"      description:"Status: 0=disabled, 1=enabled"`
	IsFrame    int         `json:"isFrame"    orm:"is_frame"    description:"External link flag: 1=yes, 0=no"`
	IsCache    int         `json:"isCache"    orm:"is_cache"    description:"Cache flag: 1=yes, 0=no"`
	QueryParam string      `json:"queryParam" orm:"query_param" description:"Route parameters in JSON format"`
	Remark     string      `json:"remark"     orm:"remark"      description:"Remark"`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:"Creation time"`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:"Update time"`
	DeletedAt  *gtime.Time `json:"deletedAt"  orm:"deleted_at"  description:"Deletion time"`
}
