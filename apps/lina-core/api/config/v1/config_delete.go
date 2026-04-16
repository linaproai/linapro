package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Config Delete API

// DeleteReq defines the request for deleting a config.
type DeleteReq struct {
	g.Meta `path:"/config/{id}" method:"delete" tags:"参数设置" summary:"删除参数设置" dc:"根据参数ID软删除指定的参数设置记录" permission:"system:config:remove"`
	Id     int `json:"id" v:"required" dc:"参数ID" eg:"1"`
}

// DeleteRes is the config delete response.
type DeleteRes struct{}
