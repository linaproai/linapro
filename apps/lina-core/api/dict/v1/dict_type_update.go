package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TypeUpdateReq defines the request for updating a dictionary type.
type TypeUpdateReq struct {
	g.Meta `path:"/dict/type/{id}" method:"put" tags:"Dictionary Management" summary:"Update dictionary type" dc:"Update the information of the specified dictionary type. Modifying the dictionary type key will update the associated dictionary data simultaneously." permission:"system:dict:edit"`
	Id     int     `json:"id" v:"required" dc:"Dictionary type ID" eg:"1"`
	Name   *string `json:"name" dc:"Dictionary name" eg:"User gender"`
	Type   *string `json:"type" dc:"Dictionary type identifier" eg:"sys_user_sex"`
	Status *int    `json:"status" dc:"Status: 1=normal 0=disabled" eg:"1"`
	Remark *string `json:"remark" dc:"Remarks" eg:"User gender options"`
}

// TypeUpdateRes defines the response for updating a dictionary type.
type TypeUpdateRes struct{}
