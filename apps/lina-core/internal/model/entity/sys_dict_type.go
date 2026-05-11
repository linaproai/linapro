// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDictType is the golang structure for table sys_dict_type.
type SysDictType struct {
	Id                  int         `json:"id"                  orm:"id"                    description:"Dictionary type ID"`
	TenantId            int         `json:"tenantId"            orm:"tenant_id"             description:"Owning tenant ID, 0 means PLATFORM default"`
	Name                string      `json:"name"                orm:"name"                  description:"Dictionary name"`
	Type                string      `json:"type"                orm:"type"                  description:"Dictionary type"`
	Status              int         `json:"status"              orm:"status"                description:"Status: 0=disabled, 1=enabled"`
	IsBuiltin           int         `json:"isBuiltin"           orm:"is_builtin"            description:"Built-in record flag: 1=yes, 0=no"`
	AllowTenantOverride bool        `json:"allowTenantOverride" orm:"allow_tenant_override" description:"Whether tenants may override this dictionary type"`
	Remark              string      `json:"remark"              orm:"remark"                description:"Remark"`
	CreatedAt           *gtime.Time `json:"createdAt"           orm:"created_at"            description:"Creation time"`
	UpdatedAt           *gtime.Time `json:"updatedAt"           orm:"updated_at"            description:"Update time"`
	DeletedAt           *gtime.Time `json:"deletedAt"           orm:"deleted_at"            description:"Deletion time"`
}
