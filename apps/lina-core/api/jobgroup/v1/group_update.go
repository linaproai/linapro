package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateReq defines the request for updating one scheduled job group.
type UpdateReq struct {
	g.Meta `path:"/job-group/{id}" method:"put" tags:"Job Scheduling / Job Groups" summary:"Update group" dc:"Update the scheduled job group information based on the group ID. The default group only allows modification of open fields." permission:"system:jobgroup:edit"`
	Id     uint64 `json:"id" v:"required" dc:"Group ID" eg:"1"`
	GroupPayload
}

// UpdateRes defines the response for updating one scheduled job group.
type UpdateRes struct{}
