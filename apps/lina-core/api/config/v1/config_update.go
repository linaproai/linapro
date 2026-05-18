package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Config Update API

// UpdateReq defines the request for updating a config.
type UpdateReq struct {
	g.Meta `path:"/config/{id}" method:"put" tags:"Parameter Settings" summary:"Update parameter settings" dc:"Update the information of the specified parameter settings. When modifying the parameter key name, make sure it does not conflict with other parameters." permission:"system:config:edit"`
	Id     int     `json:"id" v:"required" dc:"Parameter ID" eg:"1"`
	Name   *string `json:"name" dc:"Parameter name" eg:"Main frame page-default skin style name"`
	Key    *string `json:"key" dc:"Parameter key name (unique identifier)" eg:"sys.index.skinName"`
	Value  *string `json:"value" dc:"Parameter key value" eg:"skin-blue"`
	Remark *string `json:"remark" dc:"Remarks" eg:"blue skin-blue, green skin-green"`
}

// UpdateRes is the config update response.
type UpdateRes struct{}
