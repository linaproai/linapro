package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateReq defines the request for creating a post.
type CreateReq struct {
	g.Meta `path:"/post" method:"post" tags:"Position Management" summary:"Create position" dc:"Create a new position under the designated department. The position code must be unique in the system" permission:"system:post:add"`
	DeptId int    `json:"deptId" v:"required#gf.gvalid.rule.required" dc:"Department ID" eg:"100"`
	Code   string `json:"code" v:"required#gf.gvalid.rule.required" dc:"Position code (unique)" eg:"dev"`
	Name   string `json:"name" v:"required#gf.gvalid.rule.required" dc:"Position name" eg:"Development Engineer"`
	Sort   *int   `json:"sort" d:"0" dc:"Sorting number, the smaller the value, the higher the sorting is." eg:"1"`
	Status *int   `json:"status" d:"1" dc:"Status: 1=normal 0=disabled" eg:"1"`
	Remark string `json:"remark" dc:"Remark" eg:"Responsible for system development"`
}

// CreateRes defines the response for creating a post.
type CreateRes struct {
	Id int `json:"id" dc:"Position ID" eg:"1"`
}
