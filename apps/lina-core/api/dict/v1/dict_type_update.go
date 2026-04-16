package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TypeUpdateReq defines the request for updating a dictionary type.
type TypeUpdateReq struct {
	g.Meta `path:"/dict/type/{id}" method:"put" tags:"字典管理" summary:"更新字典类型" dc:"更新指定字典类型的信息，修改字典类型标识会同步更新关联的字典数据" permission:"system:dict:edit"`
	Id     int     `json:"id" v:"required" dc:"字典类型ID" eg:"1"`
	Name   *string `json:"name" dc:"字典名称" eg:"用户性别"`
	Type   *string `json:"type" dc:"字典类型标识" eg:"sys_user_sex"`
	Status *int    `json:"status" dc:"状态：1=正常 0=停用" eg:"1"`
	Remark *string `json:"remark" dc:"备注" eg:"用户性别选项"`
}

// TypeUpdateRes defines the response for updating a dictionary type.
type TypeUpdateRes struct{}
