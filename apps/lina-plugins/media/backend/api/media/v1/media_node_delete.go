// This file declares media node delete DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteNodeReq defines the request for deleting one media node.
type DeleteNodeReq struct {
	g.Meta  `path:"/media/nodes/{nodeNum}" method:"delete" tags:"媒体管理" summary:"删除节点" dc:"按节点编号删除未被引用的节点。" permission:"media:management:remove"`
	NodeNum int `json:"nodeNum" v:"min:0|max:255#节点编号不能小于0|节点编号不能大于255" dc:"节点编号" eg:"1"`
}

// DeleteNodeRes defines the response for deleting one media node.
type DeleteNodeRes struct {
	NodeNum int `json:"nodeNum" dc:"节点编号" eg:"1"`
}
