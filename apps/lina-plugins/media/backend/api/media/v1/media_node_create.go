// This file declares media node create DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateNodeReq defines the request for creating one media node.
type CreateNodeReq struct {
	g.Meta   `path:"/media/nodes" method:"post" tags:"媒体管理" summary:"新增节点" dc:"新增一个媒体节点配置。" permission:"media:management:add"`
	NodeNum  int    `json:"nodeNum" v:"min:0|max:255#节点编号不能小于0|节点编号不能大于255" dc:"节点编号" eg:"1"`
	Name     string `json:"name" v:"required|length:1,32#节点名称不能为空|节点名称长度不能超过32个字符" dc:"节点名称" eg:"华东节点"`
	QnUrl    string `json:"qnUrl" v:"required|length:1,255#节点网关地址不能为空|节点网关地址长度不能超过255个字符" dc:"节点网关地址" eg:"https://qn.example.com"`
	BasicUrl string `json:"basicUrl" v:"required|length:1,255#基础平台网关地址不能为空|基础平台网关地址长度不能超过255个字符" dc:"基础平台网关地址" eg:"https://basic.example.com"`
	DnUrl    string `json:"dnUrl" v:"required|length:1,255#属地网关地址不能为空|属地网关地址长度不能超过255个字符" dc:"属地网关地址" eg:"https://dn.example.com"`
}

// CreateNodeRes defines the response for creating one media node.
type CreateNodeRes struct {
	NodeNum int `json:"nodeNum" dc:"节点编号" eg:"1"`
}
