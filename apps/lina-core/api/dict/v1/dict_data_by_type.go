package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DataByTypeReq defines the request for querying dictionary data by type.
type DataByTypeReq struct {
	g.Meta   `path:"/dict/data/type/{dictType}" method:"get" tags:"字典管理" summary:"按类型获取字典数据" dc:"根据字典类型标识获取该类型下所有正常状态的字典数据项，供管理工作台的选择器、筛选器或表单选项复用" permission:"system:dict:query"`
	DictType string `json:"dictType" v:"required" dc:"字典类型标识" eg:"sys_user_sex"`
}

// DataByTypeRes defines the response for querying dictionary data by type.
type DataByTypeRes struct {
	List []*entity.SysDictData `json:"list" dc:"字典数据列表" eg:"[]"`
}
