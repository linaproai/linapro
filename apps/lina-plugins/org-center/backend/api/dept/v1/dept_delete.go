package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteReq defines the request for deleting a department.
type DeleteReq struct {
	g.Meta `path:"/dept/{id}" method:"delete" tags:"Department Management" summary:"Delete department" dc:"Delete the specified department. If it has child departments or associated users, remove those records before deleting it." permission:"system:dept:remove"`
	Id     int `json:"id" v:"required" dc:"Department ID to be deleted" eg:"110"`
}

// DeleteRes defines the response for deleting a department.
type DeleteRes struct{}
