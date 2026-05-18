package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DataDeleteReq defines the request for deleting dictionary data.
type DataDeleteReq struct {
	g.Meta `path:"/dict/data/{id}" method:"delete" tags:"Dictionary Management" summary:"Delete dictionary data" dc:"Delete the specified dictionary data item" permission:"system:dict:remove"`
	Id     int `json:"id" v:"required" dc:"Dictionary data ID" eg:"1"`
}

// DataDeleteRes defines the response for deleting dictionary data.
type DataDeleteRes struct{}
