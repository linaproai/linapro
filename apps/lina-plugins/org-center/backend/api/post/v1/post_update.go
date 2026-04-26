package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateReq defines the request for updating a post.
type UpdateReq struct {
	g.Meta `path:"/post/{id}" method:"put" tags:"Position Management" summary:"Update position" dc:"Update the specified position. All fields are optional." permission:"system:post:edit"`
	Id     int     `json:"id" v:"required" dc:"Position ID" eg:"1"`
	DeptId *int    `json:"deptId" dc:"Department ID" eg:"100"`
	Code   *string `json:"code" dc:"Position code (unique)" eg:"dev"`
	Name   *string `json:"name" dc:"Position name" eg:"Development Engineer"`
	Sort   *int    `json:"sort" dc:"Sort order" eg:"1"`
	Status *int    `json:"status" dc:"Status: 1=normal 0=disabled" eg:"1"`
	Remark *string `json:"remark" dc:"Remark" eg:"Responsible for system development"`
}

// UpdateRes defines the response for updating a post.
type UpdateRes struct{}
