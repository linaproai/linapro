// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MediaStrategyDevice is the golang structure of table media_strategy_device for DAO operations like Where/Data.
type MediaStrategyDevice struct {
	g.Meta     `orm:"table:media_strategy_device, do:true"`
	DeviceId   any // 设备国标ID
	StrategyId any // 策略ID
}
