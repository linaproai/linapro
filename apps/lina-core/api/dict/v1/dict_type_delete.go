package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TypeDeleteReq defines the request for deleting a dictionary type.
type TypeDeleteReq struct {
	g.Meta `path:"/dict/type/{id}" method:"delete" tags:"Dictionary Management" summary:"Delete dictionary type" dc:"Delete the specified dictionary type and delete all dictionary data under this type in cascade" permission:"system:dict:remove"`
	Id     int `json:"id" v:"required" dc:"Dictionary type ID" eg:"1"`
}

// TypeDeleteRes defines the response for deleting a dictionary type.
type TypeDeleteRes struct{}
