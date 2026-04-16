package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TypeCreateReq defines the request for creating a dictionary type.
type TypeCreateReq struct {
	g.Meta `path:"/dict/type" method:"post" tags:"字典管理" summary:"创建字典类型" dc:"创建一个新的字典类型，字典类型标识在系统中必须唯一" permission:"system:dict:add"`
	Name   string `json:"name" v:"required#请输入字典名称" dc:"字典名称" eg:"用户性别"`
	Type   string `json:"type" v:"required#请输入字典类型" dc:"字典类型标识（唯一）" eg:"sys_user_sex"`
	Status *int   `json:"status" d:"1" dc:"状态：1=正常 0=停用" eg:"1"`
	Remark string `json:"remark" dc:"备注" eg:"用户性别选项"`
}

// TypeCreateRes defines the response for creating a dictionary type.
type TypeCreateRes struct {
	Id int `json:"id" dc:"字典类型ID" eg:"1"`
}
