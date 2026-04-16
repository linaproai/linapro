package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DataCreateReq defines the request for creating dictionary data.
type DataCreateReq struct {
	g.Meta   `path:"/dict/data" method:"post" tags:"字典管理" summary:"创建字典数据" dc:"在指定字典类型下创建一条字典数据项" permission:"system:dict:add"`
	DictType string `json:"dictType" v:"required#请输入字典类型" dc:"所属字典类型标识" eg:"sys_user_sex"`
	Label    string `json:"label" v:"required#请输入字典标签" dc:"字典标签（显示名称）" eg:"男"`
	Value    string `json:"value" v:"required#请输入字典值" dc:"字典值（存储值）" eg:"1"`
	Sort     *int   `json:"sort" d:"0" dc:"排序号，数值越小排序越靠前" eg:"1"`
	TagStyle string `json:"tagStyle" dc:"标签样式，用于前端展示不同颜色标签" eg:"success"`
	CssClass string `json:"cssClass" dc:"CSS类名，用于前端自定义样式" eg:"text-green"`
	Status   *int   `json:"status" d:"1" dc:"状态：1=正常 0=停用" eg:"1"`
	Remark   string `json:"remark" dc:"备注" eg:"性别男"`
}

// DataCreateRes defines the response for creating dictionary data.
type DataCreateRes struct {
	Id int `json:"id" dc:"字典数据ID" eg:"1"`
}
