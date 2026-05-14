// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MediaStrategyTenant is the golang structure of table media_strategy_tenant for DAO operations like Where/Data.
type MediaStrategyTenant struct {
	g.Meta     `orm:"table:media_strategy_tenant, do:true"`
	TenantId   any // 租户ID
	StrategyId any // 策略ID
}
