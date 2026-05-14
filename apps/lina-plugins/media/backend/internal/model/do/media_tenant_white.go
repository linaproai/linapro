// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaTenantWhite is the golang structure of table media_tenant_white for DAO operations like Where/Data.
type MediaTenantWhite struct {
	g.Meta      `orm:"table:media_tenant_white, do:true"`
	TenantId    any         // 租户ID
	Ip          any         // 白名单地址
	Description any         // 白名单描述
	Enable      any         // 1开启，0关闭
	CreatorId   any         // 创建人ID
	CreateTime  *gtime.Time // 创建时间
	UpdaterId   any         // 修改人ID
	UpdateTime  *gtime.Time // 修改时间
}
