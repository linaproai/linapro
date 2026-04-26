package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// TypeGetReq defines the request for querying dictionary type detail.
type TypeGetReq struct {
	g.Meta `path:"/dict/type/{id}" method:"get" tags:"Dictionary Management" summary:"Get dictionary type details" dc:"Get details of dictionary type based on dictionary type ID" permission:"system:dict:query"`
	Id     int `json:"id" v:"required" dc:"Dictionary type ID" eg:"1"`
}

// TypeGetRes defines the response for querying dictionary type detail.
type TypeGetRes struct {
	*entity.SysDictType `dc:"Dictionary type information" eg:""`
}
