// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDictData is the golang structure for table sys_dict_data.
type SysDictData struct {
	Id        int         `json:"id"        orm:"id"         description:"Dictionary data ID"`
	DictType  string      `json:"dictType"  orm:"dict_type"  description:"Dictionary type"`
	Label     string      `json:"label"     orm:"label"      description:"Dictionary label"`
	Value     string      `json:"value"     orm:"value"      description:"Dictionary value"`
	Sort      int         `json:"sort"      orm:"sort"       description:"Display order"`
	TagStyle  string      `json:"tagStyle"  orm:"tag_style"  description:"Tag style: primary/success/danger/warning, etc."`
	CssClass  string      `json:"cssClass"  orm:"css_class"  description:"CSS class name"`
	Status    int         `json:"status"    orm:"status"     description:"Status: 0=disabled, 1=enabled"`
	IsBuiltin int         `json:"isBuiltin" orm:"is_builtin" description:"Built-in record flag: 1=yes, 0=no"`
	Remark    string      `json:"remark"    orm:"remark"     description:"Remark"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
	TenantId  int         `json:"tenantId"  orm:"tenant_id"  description:"Owning tenant ID, 0 means PLATFORM default"`
}
