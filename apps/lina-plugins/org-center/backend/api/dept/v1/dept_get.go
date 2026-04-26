package v1

import "github.com/gogf/gf/v2/frame/g"

// GetReq defines the request for querying department detail.
type GetReq struct {
	g.Meta `path:"/dept/{id}" method:"get" tags:"Department Management" summary:"Get department details" dc:"Get the complete details of the department based on the department ID, including basic information, owner, contact information, etc." permission:"system:dept:query"`
	Id     int `json:"id" v:"required" dc:"Department ID" eg:"100"`
}

// GetRes is the response for department detail.
type GetRes struct {
	*DeptEntity `dc:"Department details, including all field data of the department" eg:""`
}
