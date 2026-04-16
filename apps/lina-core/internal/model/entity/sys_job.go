// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJob is the golang structure for table sys_job.
type SysJob struct {
	Id          uint64      `json:"id"          orm:"id"          description:"任务ID"`
	Name        string      `json:"name"        orm:"name"        description:"任务名称"`
	Command     string      `json:"command"     orm:"command"     description:"执行指令"`
	CronExpr    string      `json:"cronExpr"    orm:"cron_expr"   description:"Cron表达式"`
	Description string      `json:"description" orm:"description" description:"任务描述"`
	Status      int         `json:"status"      orm:"status"      description:"状态：1=启用 0=禁用"`
	Singleton   int         `json:"singleton"   orm:"singleton"   description:"执行模式：1=单例 0=并行"`
	MaxTimes    int         `json:"maxTimes"    orm:"max_times"   description:"最大执行次数，0表示无限制"`
	ExecTimes   int         `json:"execTimes"   orm:"exec_times"  description:"已执行次数"`
	IsSystem    int         `json:"isSystem"    orm:"is_system"   description:"是否系统任务：1=是 0=否"`
	CreateBy    string      `json:"createBy"    orm:"create_by"   description:"创建者"`
	CreateTime  *gtime.Time `json:"createTime"  orm:"create_time" description:"创建时间"`
	UpdateBy    string      `json:"updateBy"    orm:"update_by"   description:"更新者"`
	UpdateTime  *gtime.Time `json:"updateTime"  orm:"update_time" description:"更新时间"`
	Remark      string      `json:"remark"      orm:"remark"      description:"备注"`
}
