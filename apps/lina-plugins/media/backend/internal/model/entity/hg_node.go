// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// HgNode is the golang structure for table hg_node.
type HgNode struct {
	Id         int         `json:"id"         orm:"id"          description:"ID（自增，无符号）"`
	NodeNum    int         `json:"nodeNum"    orm:"node_num"    description:"节点编号"`
	Name       string      `json:"name"       orm:"name"        description:"节点名称"`
	QnUrl      string      `json:"qnUrl"      orm:"qn_url"      description:"节点网关地址"`
	BasicUrl   string      `json:"basicUrl"   orm:"basic_url"   description:"基础平台网关地址"`
	DnUrl      string      `json:"dnUrl"      orm:"dn_url"      description:"属地网关地址"`
	CreatorId  int         `json:"creatorId"  orm:"creator_id"  description:"创建人ID"`
	CreateTime *gtime.Time `json:"createTime" orm:"create_time" description:"创建时间"`
	UpdaterId  int         `json:"updaterId"  orm:"updater_id"  description:"修改人ID"`
	UpdateTime *gtime.Time `json:"updateTime" orm:"update_time" description:"修改时间"`
}
