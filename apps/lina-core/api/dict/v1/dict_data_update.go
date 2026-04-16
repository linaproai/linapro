package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DataUpdateReq defines the request for updating dictionary data.
type DataUpdateReq struct {
	g.Meta   `path:"/dict/data/{id}" method:"put" tags:"字典管理" summary:"更新字典数据" dc:"更新指定字典数据项的信息" permission:"system:dict:edit"`
	Id       int     `json:"id" v:"required" dc:"字典数据ID" eg:"1"`
	DictType *string `json:"dictType" dc:"所属字典类型标识" eg:"sys_user_sex"`
	Label    *string `json:"label" dc:"字典标签（显示名称）" eg:"男"`
	Value    *string `json:"value" dc:"字典值（存储值）" eg:"1"`
	Sort     *int    `json:"sort" dc:"排序号" eg:"1"`
	TagStyle *string `json:"tagStyle" dc:"标签样式" eg:"success"`
	CssClass *string `json:"cssClass" dc:"CSS类名" eg:"text-green"`
	Status   *int    `json:"status" dc:"状态：1=正常 0=停用" eg:"1"`
	Remark   *string `json:"remark" dc:"备注" eg:"性别男"`
}

// DataUpdateRes defines the response for updating dictionary data.
type DataUpdateRes struct{}
