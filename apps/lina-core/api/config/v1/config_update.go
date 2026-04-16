package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Config Update API

// UpdateReq defines the request for updating a config.
type UpdateReq struct {
	g.Meta `path:"/config/{id}" method:"put" tags:"参数设置" summary:"更新参数设置" dc:"更新指定参数设置的信息，修改参数键名时需确保不与其他参数冲突" permission:"system:config:edit"`
	Id     int     `json:"id" v:"required" dc:"参数ID" eg:"1"`
	Name   *string `json:"name" dc:"参数名称" eg:"主框架页-默认皮肤样式名称"`
	Key    *string `json:"key" dc:"参数键名（唯一标识）" eg:"sys.index.skinName"`
	Value  *string `json:"value" dc:"参数键值" eg:"skin-blue"`
	Remark *string `json:"remark" dc:"备注" eg:"蓝色 skin-blue、绿色 skin-green"`
}

// UpdateRes is the config update response.
type UpdateRes struct{}
