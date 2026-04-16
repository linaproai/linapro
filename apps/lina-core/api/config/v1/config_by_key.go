package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Config ByKey API

// ByKeyReq defines the request for getting config by key name.
type ByKeyReq struct {
	g.Meta `path:"/config/key/{key}" method:"get" tags:"参数设置" summary:"按键名获取参数" dc:"根据参数键名获取参数设置的详细信息，用于其他模块按键名查询参数值" permission:"system:config:query"`
	Key    string `json:"key" v:"required" dc:"参数键名" eg:"sys.user.initPassword"`
}

// ByKeyRes is the response for getting config by key.
type ByKeyRes struct {
	*entity.SysConfig `dc:"参数设置信息" eg:""`
}
