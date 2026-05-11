// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserMembership is the golang structure for table user_membership.
type UserMembership struct {
	Id        int64       `json:"id"        orm:"id"         description:""`
	UserId    int64       `json:"userId"    orm:"user_id"    description:""`
	TenantId  int64       `json:"tenantId"  orm:"tenant_id"  description:""`
	Status    int         `json:"status"    orm:"status"     description:""`
	JoinedAt  *gtime.Time `json:"joinedAt"  orm:"joined_at"  description:""`
	CreatedBy int64       `json:"createdBy" orm:"created_by" description:""`
	UpdatedBy int64       `json:"updatedBy" orm:"updated_by" description:""`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
