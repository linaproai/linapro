// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysRole is the golang structure for table sys_role.
type SysRole struct {
	Id        int         `json:"id"        orm:"id"         description:"角色ID"`
	Name      string      `json:"name"      orm:"name"       description:"角色名称"`
	Key       string      `json:"key"       orm:"key"        description:"权限字符"`
	Sort      int         `json:"sort"      orm:"sort"       description:"显示排序"`
	DataScope int         `json:"dataScope" orm:"data_scope" description:"数据权限范围（1=全部 2=本部门 3=仅本人）"`
	Status    int         `json:"status"    orm:"status"     description:"状态（0=停用 1=正常）"`
	Remark    string      `json:"remark"    orm:"remark"     description:"备注"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"删除时间"`
}
