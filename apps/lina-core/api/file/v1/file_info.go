package v1

import "github.com/gogf/gf/v2/frame/g"

// InfoByIdsReq defines the request for querying file info by IDs.
type InfoByIdsReq struct {
	g.Meta `path:"/file/info/{ids}" method:"get" tags:"File Management" summary:"Query file information based on ID" dc:"Query file details based on file ID, support batch query (comma separated multiple IDs), used for file echo" permission:"system:file:query"`
	Ids    string `json:"ids" v:"required" dc:"File ID, multiple separated by commas" eg:"1,2,3"`
}

// InfoByIdsRes defines the response for file info queries.
type InfoByIdsRes struct {
	List []*FileItem `json:"list" dc:"File information list" eg:"[]"`
}
