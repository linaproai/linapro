// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaStrategy is the golang structure of table media_strategy for DAO operations like Where/Data.
type MediaStrategy struct {
	g.Meta     `orm:"table:media_strategy, do:true"`
	Id         any         // 策略ID
	Name       any         // 策略名称
	Strategy   any         // YAML格式策略内容
	Global     any         // 是否全局策略：1是，2否
	Enable     any         // 启用状态：1开启，2关闭
	CreatorId  any         // 创建人ID
	UpdaterId  any         // 修改人ID
	CreateTime *gtime.Time // 创建时间
	UpdateTime *gtime.Time // 修改时间
}
