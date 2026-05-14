// This file declares media node list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListNodesReq defines the request for querying media nodes.
type ListNodesReq struct {
	g.Meta   `path:"/media/nodes" method:"get" tags:"媒体管理" summary:"查询节点列表" dc:"分页查询节点配置，支持按节点名称、节点编号或网关地址模糊筛选。" permission:"media:management:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Keyword  string `json:"keyword" dc:"按节点编号、名称或网关地址模糊筛选" eg:"1"`
}

// ListNodesRes defines the response for querying media nodes.
type ListNodesRes struct {
	List  []*NodeListItem `json:"list" dc:"节点列表" eg:"[]"`
	Total int             `json:"total" dc:"匹配总数" eg:"1"`
}

// NodeListItem defines one media node row.
type NodeListItem struct {
	Id         int    `json:"id" dc:"ID" eg:"1"`
	NodeNum    int    `json:"nodeNum" dc:"节点编号" eg:"1"`
	Name       string `json:"name" dc:"节点名称" eg:"华东节点"`
	QnUrl      string `json:"qnUrl" dc:"节点网关地址" eg:"https://qn.example.com"`
	BasicUrl   string `json:"basicUrl" dc:"基础平台网关地址" eg:"https://basic.example.com"`
	DnUrl      string `json:"dnUrl" dc:"属地网关地址" eg:"https://dn.example.com"`
	CreatorId  int    `json:"creatorId" dc:"创建人ID" eg:"1"`
	CreateTime string `json:"createTime" dc:"创建时间" eg:"2026-05-13 10:00:00"`
	UpdaterId  int    `json:"updaterId" dc:"修改人ID" eg:"1"`
	UpdateTime string `json:"updateTime" dc:"修改时间" eg:"2026-05-13 10:00:00"`
}
