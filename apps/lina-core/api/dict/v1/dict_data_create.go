// This file defines the dictionary-data creation DTOs and validation rules.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DataCreateReq defines the request for creating dictionary data.
type DataCreateReq struct {
	g.Meta   `path:"/dict/data" method:"post" tags:"Dictionary Management" summary:"Create dictionary data" dc:"Create a dictionary data item under the specified dictionary type" permission:"system:dict:add"`
	DictType string `json:"dictType" v:"required#validation.dict.data.create.dictType.required" dc:"Dictionary type identifier" eg:"sys_user_sex"`
	Label    string `json:"label" v:"required#validation.dict.data.create.label.required" dc:"dictionary tag (display name)" eg:"male"`
	Value    string `json:"value" v:"required#validation.dict.data.create.value.required" dc:"Dictionary value (stored value)" eg:"1"`
	Sort     *int   `json:"sort" d:"0" dc:"Sorting number, the smaller the value, the higher the sorting is." eg:"1"`
	TagStyle string `json:"tagStyle" dc:"Label style, used to display labels of different colors on the front end" eg:"success"`
	CssClass string `json:"cssClass" dc:"CSS class name, used for frontend custom styles" eg:"text-green"`
	Status   *int   `json:"status" d:"1" dc:"Status: 1=normal 0=disabled" eg:"1"`
	Remark   string `json:"remark" dc:"Remarks" eg:"Gender male"`
}

// DataCreateRes defines the response for creating dictionary data.
type DataCreateRes struct {
	Id int `json:"id" dc:"Dictionary data ID" eg:"1"`
}
