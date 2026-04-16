package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Config Get API

// GetReq defines the request for getting config detail by ID.
type GetReq struct {
	g.Meta `path:"/config/{id}" method:"get" tags:"参数设置" summary:"获取参数设置详情" dc:"根据参数ID获取参数设置的详细信息" permission:"system:config:query"`
	Id     int `json:"id" v:"required" dc:"参数ID" eg:"1"`
}

// GetRes is the config detail response.
type GetRes struct {
	*entity.SysConfig `dc:"参数设置信息" eg:""`
}
