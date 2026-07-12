package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteReq defines the request for deleting scheduled job groups.
type DeleteReq struct {
	g.Meta `path:"/job-group" method:"delete" tags:"Job Scheduling / Group Management" summary:"Delete group" dc:"Delete task groups in batches by group ID. Delete is not allowed in the default group. Tasks in other groups need to be migrated to the default group. Pass ids as a query array (ids[]=2&ids[]=3)." permission:"system:jobgroup:remove"`
	Ids    []int64 `json:"ids" v:"required|min-length:1" dc:"Group ID list as a query array, e.g. ids[]=2&ids[]=3" eg:"[2,3]"`
}

// DeleteRes defines the response for deleting scheduled job groups.
type DeleteRes struct{}
