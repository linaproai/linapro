// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Operlog is the golang structure of table plugin_monitor_operlog for DAO operations like Where/Data.
type Operlog struct {
	g.Meta        `orm:"table:plugin_monitor_operlog, do:true"`
	Id            any         // 日志ID
	Title         any         // 模块标题
	OperSummary   any         // 操作摘要
	OperType      any         // 操作类型（create新增 update修改 delete删除 export导出 import导入 other其他）
	Method        any         // 方法名称
	RequestMethod any         // 请求方式（GET/POST/PUT/DELETE）
	OperName      any         // 操作人员
	OperUrl       any         // 请求URL
	OperIp        any         // 操作IP地址
	OperParam     any         // 请求参数
	JsonResult    any         // 返回参数
	Status        any         // 操作状态（0成功 1失败）
	ErrorMsg      any         // 错误消息
	CostTime      any         // 耗时（毫秒）
	OperTime      *gtime.Time // 操作时间
}
