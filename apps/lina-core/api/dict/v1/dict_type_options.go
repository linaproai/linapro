package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// TypeOptionsReq defines the request for querying dictionary type options.
type TypeOptionsReq struct {
	g.Meta `path:"/dict/type/options" method:"get" tags:"字典管理" summary:"获取全部字典类型选项" dc:"获取所有正常状态的字典类型列表，供管理工作台的字典维护视图或其他类型选择器装配选项" permission:"system:dict:query"`
}

// TypeOptionsRes defines the response for querying dictionary type options.
type TypeOptionsRes struct {
	List []*entity.SysDictType `json:"list" dc:"字典类型选项列表" eg:"[]"`
}
