package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TypeDeleteReq defines the request for deleting a dictionary type.
type TypeDeleteReq struct {
	g.Meta `path:"/dict/type/{id}" method:"delete" tags:"字典管理" summary:"删除字典类型" dc:"删除指定字典类型，同时级联删除该类型下的所有字典数据" permission:"system:dict:remove"`
	Id     int `json:"id" v:"required" dc:"字典类型ID" eg:"1"`
}

// TypeDeleteRes defines the response for deleting a dictionary type.
type TypeDeleteRes struct{}
