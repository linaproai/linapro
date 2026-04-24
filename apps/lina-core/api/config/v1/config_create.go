// This file defines the configuration-creation DTOs and validation rules.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Config Create API

// CreateReq defines the request for creating a config.
type CreateReq struct {
	g.Meta `path:"/config" method:"post" tags:"参数设置" summary:"创建参数设置" dc:"创建一个新的参数设置，参数键名在系统中必须唯一" permission:"system:config:add"`
	Name   string `json:"name" v:"required#validation.config.create.name.required" dc:"参数名称" eg:"主框架页-默认皮肤样式名称"`
	Key    string `json:"key" v:"required#validation.config.create.key.required" dc:"参数键名（唯一标识）" eg:"sys.index.skinName"`
	Value  string `json:"value" v:"required#validation.config.create.value.required" dc:"参数键值" eg:"skin-blue"`
	Remark string `json:"remark" dc:"备注" eg:"蓝色 skin-blue、绿色 skin-green"`
}

// CreateRes is the config creation response.
type CreateRes struct {
	Id int `json:"id" dc:"参数ID" eg:"1"`
}
