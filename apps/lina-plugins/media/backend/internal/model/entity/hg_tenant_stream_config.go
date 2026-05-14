// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// HgTenantStreamConfig is the golang structure for table hg_tenant_stream_config.
type HgTenantStreamConfig struct {
	TenantId      string      `json:"tenantId"      orm:"tenant_id"      description:"租户ID"`
	MaxConcurrent int         `json:"maxConcurrent" orm:"max_concurrent" description:"最大并发数"`
	NodeNum       int         `json:"nodeNum"       orm:"node_num"       description:"节点编号"`
	Enable        int         `json:"enable"        orm:"enable"         description:"1开启，0关闭"`
	CreatorId     int         `json:"creatorId"     orm:"creator_id"     description:"创建人ID"`
	CreateTime    *gtime.Time `json:"createTime"    orm:"create_time"    description:"创建时间"`
	UpdaterId     int         `json:"updaterId"     orm:"updater_id"     description:"修改人ID"`
	UpdateTime    *gtime.Time `json:"updateTime"    orm:"update_time"    description:"修改时间"`
}
