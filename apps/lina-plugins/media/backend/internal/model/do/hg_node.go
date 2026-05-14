// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// HgNode is the golang structure of table hg_node for DAO operations like Where/Data.
type HgNode struct {
	g.Meta     `orm:"table:hg_node, do:true"`
	Id         any         // ID（自增，无符号）
	NodeNum    any         // 节点编号
	Name       any         // 节点名称
	QnUrl      any         // 节点网关地址
	BasicUrl   any         // 基础平台网关地址
	DnUrl      any         // 属地网关地址
	CreatorId  any         // 创建人ID
	CreateTime *gtime.Time // 创建时间
	UpdaterId  any         // 修改人ID
	UpdateTime *gtime.Time // 修改时间
}
