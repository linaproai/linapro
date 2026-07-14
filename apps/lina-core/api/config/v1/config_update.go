package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/configvaluetype"
)

// Config Update API

// UpdateReq defines the request for updating a config.
type UpdateReq struct {
	g.Meta    `path:"/config/{id}" method:"put" tags:"Parameter Settings" summary:"Update parameter settings" dc:"Update the information of the specified parameter settings. When modifying the parameter key name, make sure it does not conflict with other parameters." permission:"system:config:edit"`
	Id        int64                 `json:"id" v:"required" dc:"Parameter ID" eg:"1"`
	Name      *string               `json:"name" dc:"Parameter name" eg:"Main frame page-default skin style name"`
	Key       *string               `json:"key" dc:"Parameter key name (unique identifier)" eg:"sys.index.skinName"`
	Value     *string               `json:"value" dc:"Parameter key value" eg:"skin-blue"`
	ValueType *configvaluetype.Code `json:"valueType" dc:"Value input type: text, textarea, number, boolean, select, radio, multi_select, richtext; built-in parameters cannot change type" eg:"select"`
	Options   *[]ConfigValueOption  `json:"options" dc:"Selectable options for select/radio/multi_select; built-in parameters cannot change options" eg:"[{\"label\":\"Blue\",\"value\":\"skin-blue\"}]"`
	Remark    *string               `json:"remark" dc:"Remarks" eg:"blue skin-blue, green skin-green"`
}

// UpdateRes is the config update response.
type UpdateRes struct{}
