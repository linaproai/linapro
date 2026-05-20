package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteReq defines the request for deleting scheduled job groups.
type DeleteReq struct {
	g.Meta `path:"/job-group/{ids}" method:"delete" tags:"Job Scheduling / Job Groups" summary:"Delete group" dc:"Delete task groups in batches by group ID. Delete is not allowed in the default group. Tasks in other groups need to be migrated to the default group." permission:"system:jobgroup:remove"`
	Ids    string `json:"ids" v:"required" dc:"Group ID, multiple separated by commas" eg:"2,3"`
}

// DeleteRes defines the response for deleting scheduled job groups.
type DeleteRes struct{}
