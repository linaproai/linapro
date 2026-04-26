package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeleteReq defines the request for deleting files.
type DeleteReq struct {
	g.Meta `path:"/file/{ids}" method:"delete" tags:"File Management" summary:"Delete files" dc:"Delete files based on file ID and support batch deletion (comma separated multiple IDs). Delete both physical files and database records simultaneously" permission:"system:file:remove"`
	Ids    string `json:"ids" v:"required" dc:"File ID, multiple separated by commas" eg:"1,2,3"`
}

// DeleteRes File delete response
type DeleteRes struct{}
