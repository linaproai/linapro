package v1

import "github.com/gogf/gf/v2/frame/g"

// GroupPayload defines the shared mutable fields for scheduled job groups.
type GroupPayload struct {
	Code      string `json:"code" v:"required|length:1,64" dc:"Group encoding, globally unique" eg:"default"`
	Name      string `json:"name" v:"required|length:1,128" dc:"Group name" eg:"Default grouping"`
	Remark    string `json:"remark" dc:"Group notes" eg:"System default task grouping"`
	SortOrder int    `json:"sortOrder" d:"0" dc:"Display sorting, the smaller the value, the higher it is" eg:"0"`
}

// CreateReq defines the request for creating one scheduled job group.
type CreateReq struct {
	g.Meta `path:"/job-group" method:"post" tags:"Job Scheduling / Job Groups" summary:"Create group" dc:"Create a new scheduled job group for the task management page to organize tasks by group" permission:"system:jobgroup:add"`
	GroupPayload
}

// CreateRes defines the response for creating one scheduled job group.
type CreateRes struct {
	Id int64 `json:"id" dc:"Create new group ID" eg:"1"`
}
