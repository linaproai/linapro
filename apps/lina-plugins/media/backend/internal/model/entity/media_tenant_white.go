// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaTenantWhite is the golang structure for table media_tenant_white.
type MediaTenantWhite struct {
	TenantId    string      `json:"tenantId"    orm:"tenant_id"   description:"租户ID"`
	Ip          string      `json:"ip"          orm:"ip"          description:"白名单地址"`
	Description string      `json:"description" orm:"description" description:"白名单描述"`
	Enable      int         `json:"enable"      orm:"enable"      description:"1开启，0关闭"`
	CreatorId   int         `json:"creatorId"   orm:"creator_id"  description:"创建人ID"`
	CreateTime  *gtime.Time `json:"createTime"  orm:"create_time" description:"创建时间"`
	UpdaterId   int         `json:"updaterId"   orm:"updater_id"  description:"修改人ID"`
	UpdateTime  *gtime.Time `json:"updateTime"  orm:"update_time" description:"修改时间"`
}
