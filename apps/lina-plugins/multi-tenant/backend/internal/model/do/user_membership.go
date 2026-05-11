// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserMembership is the golang structure of table plugin_multi_tenant_user_membership for DAO operations like Where/Data.
type UserMembership struct {
	g.Meta    `orm:"table:plugin_multi_tenant_user_membership, do:true"`
	Id        any         //
	UserId    any         //
	TenantId  any         //
	Status    any         //
	JoinedAt  *gtime.Time //
	CreatedBy any         //
	UpdatedBy any         //
	CreatedAt *gtime.Time //
	UpdatedAt *gtime.Time //
	DeletedAt *gtime.Time //
}
