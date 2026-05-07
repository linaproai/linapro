package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DataByTypeReq defines the request for querying dictionary data by type.
type DataByTypeReq struct {
	g.Meta   `path:"/dict/data/type/{dictType}" method:"get" tags:"Dictionary Management" summary:"Get dictionary data by type" dc:"Obtain all normal dictionary data items of this type based on the dictionary type key for reuse by selectors, filters, display tags, and form options across authenticated management workbench modules without granting dictionary management access."`
	DictType string `json:"dictType" v:"required" dc:"Dictionary type identifier" eg:"sys_user_sex"`
}

// DataByTypeRes defines the response for querying dictionary data by type.
type DataByTypeRes struct {
	List []*entity.SysDictData `json:"list" dc:"Dictionary data list" eg:"[]"`
}
