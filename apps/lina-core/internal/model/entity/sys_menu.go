// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysMenu is the golang structure for table sys_menu.
type SysMenu struct {
	Id         int         `json:"id"         orm:"id"          description:"菜单ID"`
	ParentId   int         `json:"parentId"   orm:"parent_id"   description:"父菜单ID（0=根菜单）"`
	MenuKey    string      `json:"menuKey"    orm:"menu_key"    description:"菜单稳定业务标识"`
	Name       string      `json:"name"       orm:"name"        description:"菜单名称（支持i18n）"`
	Path       string      `json:"path"       orm:"path"        description:"路由地址"`
	Component  string      `json:"component"  orm:"component"   description:"组件路径"`
	Perms      string      `json:"perms"      orm:"perms"       description:"权限标识"`
	Icon       string      `json:"icon"       orm:"icon"        description:"菜单图标"`
	Type       string      `json:"type"       orm:"type"        description:"菜单类型（D=目录 M=菜单 B=按钮）"`
	Sort       int         `json:"sort"       orm:"sort"        description:"显示排序"`
	Visible    int         `json:"visible"    orm:"visible"     description:"是否显示（1=显示 0=隐藏）"`
	Status     int         `json:"status"     orm:"status"      description:"状态（0=停用 1=正常）"`
	IsFrame    int         `json:"isFrame"    orm:"is_frame"    description:"是否外链（1=是 0=否）"`
	IsCache    int         `json:"isCache"    orm:"is_cache"    description:"是否缓存（1=是 0=否）"`
	QueryParam string      `json:"queryParam" orm:"query_param" description:"路由参数（JSON格式）"`
	Remark     string      `json:"remark"     orm:"remark"      description:"备注"`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:"创建时间"`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:"更新时间"`
	DeletedAt  *gtime.Time `json:"deletedAt"  orm:"deleted_at"  description:"删除时间"`
}
