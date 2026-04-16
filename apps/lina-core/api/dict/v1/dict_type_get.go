package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// TypeGetReq defines the request for querying dictionary type detail.
type TypeGetReq struct {
	g.Meta `path:"/dict/type/{id}" method:"get" tags:"字典管理" summary:"获取字典类型详情" dc:"根据字典类型ID获取字典类型的详细信息" permission:"system:dict:query"`
	Id     int `json:"id" v:"required" dc:"字典类型ID" eg:"1"`
}

// TypeGetRes defines the response for querying dictionary type detail.
type TypeGetRes struct {
	*entity.SysDictType `dc:"字典类型信息" eg:""`
}
