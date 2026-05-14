// This file declares media node detail DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetNodeReq defines the request for querying one media node.
type GetNodeReq struct {
	g.Meta  `path:"/media/nodes/{nodeNum}" method:"get" tags:"媒体管理" summary:"查询节点详情" dc:"按节点编号查询节点详情。" permission:"media:management:query"`
	NodeNum int `json:"nodeNum" v:"min:0|max:255#节点编号不能小于0|节点编号不能大于255" dc:"节点编号" eg:"1"`
}

// GetNodeRes defines the response for querying one media node.
type GetNodeRes = NodeListItem
