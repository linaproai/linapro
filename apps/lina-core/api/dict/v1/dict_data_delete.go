package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DataDeleteReq defines the request for deleting dictionary data.
type DataDeleteReq struct {
	g.Meta `path:"/dict/data/{id}" method:"delete" tags:"字典管理" summary:"删除字典数据" dc:"删除指定的字典数据项" permission:"system:dict:remove"`
	Id     int `json:"id" v:"required" dc:"字典数据ID" eg:"1"`
}

// DataDeleteRes defines the response for deleting dictionary data.
type DataDeleteRes struct{}
