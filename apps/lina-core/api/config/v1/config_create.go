// This file defines the configuration-creation DTOs and validation rules.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Config Create API

// CreateReq defines the request for creating a config.
type CreateReq struct {
	g.Meta `path:"/config" method:"post" tags:"Parameter Settings" summary:"Create parameter settings" dc:"Create a new parameter setting. The parameter key name must be unique in the system." permission:"system:config:add"`
	Name   string `json:"name" v:"required#validation.config.create.name.required" dc:"Parameter name" eg:"Main frame page-default skin style name"`
	Key    string `json:"key" v:"required#validation.config.create.key.required" dc:"Parameter key name (unique identifier)" eg:"sys.index.skinName"`
	Value  string `json:"value" v:"required#validation.config.create.value.required" dc:"Parameter key value" eg:"skin-blue"`
	Remark string `json:"remark" dc:"Remarks" eg:"blue skin-blue, green skin-green"`
}

// CreateRes is the config creation response.
type CreateRes struct {
	Id int `json:"id" dc:"Parameter ID" eg:"1"`
}
