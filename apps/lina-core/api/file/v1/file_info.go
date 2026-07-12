package v1

import "github.com/gogf/gf/v2/frame/g"

// InfoByIdsReq defines the request for querying file info by IDs.
type InfoByIdsReq struct {
	g.Meta `path:"/file/info" method:"get" tags:"File Management" summary:"Query file information based on ID" dc:"Query file details by ID list as a query array (ids[]=1&ids[]=2), used for file echo." permission:"system:file:query"`
	Ids    []int64 `json:"ids" v:"required|min-length:1" dc:"File ID list as a query array, e.g. ids[]=1&ids[]=2&ids[]=3" eg:"[1,2,3]"`
}

// InfoByIdsRes defines the response for file info queries.
type InfoByIdsRes struct {
	List []*FileItem `json:"list" dc:"File information list" eg:"[]"`
}
