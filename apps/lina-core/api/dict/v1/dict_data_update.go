package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DataUpdateReq defines the request for updating dictionary data.
type DataUpdateReq struct {
	g.Meta   `path:"/dict/data/{id}" method:"put" tags:"Dictionary Management" summary:"Update dictionary data" dc:"Update the information of the specified dictionary data item" permission:"system:dict:edit"`
	Id       int     `json:"id" v:"required" dc:"Dictionary data ID" eg:"1"`
	DictType *string `json:"dictType" dc:"Dictionary type identifier" eg:"sys_user_sex"`
	Label    *string `json:"label" dc:"dictionary tag (display name)" eg:"male"`
	Value    *string `json:"value" dc:"Dictionary value (stored value)" eg:"1"`
	Sort     *int    `json:"sort" dc:"sequence number" eg:"1"`
	TagStyle *string `json:"tagStyle" dc:"label style" eg:"success"`
	CssClass *string `json:"cssClass" dc:"CSS class name" eg:"text-green"`
	Status   *int    `json:"status" dc:"Status: 1=normal 0=disabled" eg:"1"`
	Remark   *string `json:"remark" dc:"Remarks" eg:"Gender male"`
}

// DataUpdateRes defines the response for updating dictionary data.
type DataUpdateRes struct{}
