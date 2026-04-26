// This file defines the dictionary-type creation DTOs and validation rules.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TypeCreateReq defines the request for creating a dictionary type.
type TypeCreateReq struct {
	g.Meta `path:"/dict/type" method:"post" tags:"Dictionary Management" summary:"Create dictionary type" dc:"Create a new dictionary type. The dictionary type key must be unique in the system" permission:"system:dict:add"`
	Name   string `json:"name" v:"required#validation.dict.type.create.name.required" dc:"Dictionary name" eg:"User gender"`
	Type   string `json:"type" v:"required#validation.dict.type.create.type.required" dc:"Dictionary type identifier (unique)" eg:"sys_user_sex"`
	Status *int   `json:"status" d:"1" dc:"Status: 1=normal 0=disabled" eg:"1"`
	Remark string `json:"remark" dc:"Remarks" eg:"User gender options"`
}

// TypeCreateRes defines the response for creating a dictionary type.
type TypeCreateRes struct {
	Id int `json:"id" dc:"Dictionary type ID" eg:"1"`
}
