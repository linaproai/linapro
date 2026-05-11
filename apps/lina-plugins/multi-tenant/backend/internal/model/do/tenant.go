// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Tenant is the golang structure of table plugin_multi_tenant_tenant for DAO operations like Where/Data.
type Tenant struct {
	g.Meta    `orm:"table:plugin_multi_tenant_tenant, do:true"`
	Id        any         //
	Code      any         //
	Name      any         //
	Status    any         //
	Remark    any         //
	CreatedBy any         //
	UpdatedBy any         //
	CreatedAt *gtime.Time //
	UpdatedAt *gtime.Time //
	DeletedAt *gtime.Time //
}
