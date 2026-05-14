// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// HgTenantStreamConfig is the golang structure of table hg_tenant_stream_config for DAO operations like Where/Data.
type HgTenantStreamConfig struct {
	g.Meta        `orm:"table:hg_tenant_stream_config, do:true"`
	TenantId      any         // 租户ID
	MaxConcurrent any         // 最大并发数
	NodeNum       any         // 节点编号
	Enable        any         // 1开启，0关闭
	CreatorId     any         // 创建人ID
	CreateTime    *gtime.Time // 创建时间
	UpdaterId     any         // 修改人ID
	UpdateTime    *gtime.Time // 修改时间
}
