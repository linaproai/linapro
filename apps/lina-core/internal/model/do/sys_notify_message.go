// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysNotifyMessage is the golang structure of table sys_notify_message for DAO operations like Where/Data.
type SysNotifyMessage struct {
	g.Meta       `orm:"table:sys_notify_message, do:true"`
	Id           any         // 主键ID
	PluginId     any         // 来源插件ID，宿主内建流程为空
	SourceType   any         // 来源类型：notice=公告 plugin=插件 system=系统
	SourceId     any         // 来源业务ID
	CategoryCode any         // 消息分类：notice=通知 announcement=公告 other=其他
	Title        any         // 消息标题
	Content      any         // 消息正文
	PayloadJson  any         // 扩展载荷JSON
	SenderUserId any         // 发送者用户ID
	CreatedAt    *gtime.Time // 创建时间
}
