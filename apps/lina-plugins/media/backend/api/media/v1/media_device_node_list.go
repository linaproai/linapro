// This file declares media device-node list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListDeviceNodesReq defines the request for querying device-node mappings.
type ListDeviceNodesReq struct {
	g.Meta   `path:"/media/device-nodes" method:"get" tags:"媒体管理" summary:"查询设备节点列表" dc:"分页查询设备节点关系，支持按设备国标ID或节点编号模糊筛选。" permission:"media:management:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Keyword  string `json:"keyword" dc:"按设备国标ID或节点编号模糊筛选" eg:"34020000001320000001"`
}

// ListDeviceNodesRes defines the response for querying device-node mappings.
type ListDeviceNodesRes struct {
	List  []*DeviceNodeListItem `json:"list" dc:"设备节点列表" eg:"[]"`
	Total int                   `json:"total" dc:"匹配总数" eg:"1"`
}

// DeviceNodeListItem defines one device-node row.
type DeviceNodeListItem struct {
	DeviceId string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000001"`
	NodeNum  int    `json:"nodeNum" dc:"节点编号" eg:"1"`
	NodeName string `json:"nodeName" dc:"节点名称" eg:"华东节点"`
}
