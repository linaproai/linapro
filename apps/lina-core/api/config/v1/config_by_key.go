package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Config ByKey API

// ByKeyReq defines the request for getting config by key name.
type ByKeyReq struct {
	g.Meta `path:"/config/key/{key}" method:"get" tags:"Parameter Settings" summary:"Get parameters by key name" dc:"Obtain detailed information about parameter settings based on the parameter key name, which can be used to query parameter values by key name in other modules." permission:"system:config:query"`
	Key    string `json:"key" v:"required" dc:"Parameter key name" eg:"sys.jwt.expire"`
}

// ByKeyRes is the response for getting config by key.
type ByKeyRes struct {
	*entity.SysConfig `dc:"Parameter setting information" eg:""`
}
