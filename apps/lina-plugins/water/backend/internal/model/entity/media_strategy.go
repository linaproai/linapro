// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaStrategy is the golang structure for table media_strategy.
type MediaStrategy struct {
	Id         int64       `json:"id"         orm:"id"          description:"策略ID"`
	Name       string      `json:"name"       orm:"name"        description:"策略名称"`
	Strategy   string      `json:"strategy"   orm:"strategy"    description:"YAML格式策略内容"`
	Global     int         `json:"global"     orm:"global"      description:"是否全局策略：1是，2否"`
	Enable     int         `json:"enable"     orm:"enable"      description:"启用状态：1开启，2关闭"`
	CreatorId  int64       `json:"creatorId"  orm:"creator_id"  description:"创建人ID"`
	UpdaterId  int64       `json:"updaterId"  orm:"updater_id"  description:"修改人ID"`
	CreateTime *gtime.Time `json:"createTime" orm:"create_time" description:"创建时间"`
	UpdateTime *gtime.Time `json:"updateTime" orm:"update_time" description:"修改时间"`
}
