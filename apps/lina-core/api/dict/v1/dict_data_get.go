package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DataGetReq defines the request for querying dictionary data detail.
type DataGetReq struct {
	g.Meta `path:"/dict/data/{id}" method:"get" tags:"字典管理" summary:"获取字典数据详情" dc:"根据字典数据ID获取字典数据项的详细信息" permission:"system:dict:query"`
	Id     int `json:"id" v:"required" dc:"字典数据ID" eg:"1"`
}

// DataGetRes defines the response for querying dictionary data detail.
type DataGetRes struct {
	*entity.SysDictData `dc:"字典数据信息" eg:""`
}
