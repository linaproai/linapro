package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeleteReq defines the request for deleting files.
type DeleteReq struct {
	g.Meta `path:"/file" method:"delete" tags:"File Management" summary:"Delete files" dc:"Delete files based on file ID list as a query array (ids[]=1&ids[]=2). Delete both physical files and database records simultaneously." permission:"system:file:remove"`
	Ids    []int64 `json:"ids" v:"required|min-length:1" dc:"File ID list as a query array, e.g. ids[]=1&ids[]=2&ids[]=3" eg:"[1,2,3]"`
}

// DeleteRes File delete response
type DeleteRes struct{}
