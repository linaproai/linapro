package v1

import "github.com/gogf/gf/v2/frame/g"

// DataGetReq defines the request for querying dictionary data detail.
type DataGetReq struct {
	g.Meta `path:"/dict/data/{id}" method:"get" tags:"Dictionary Management" summary:"Get dictionary data details" dc:"Get detailed information of dictionary data items based on dictionary data ID" permission:"system:dict:query"`
	Id     int `json:"id" v:"required" dc:"Dictionary data ID" eg:"1"`
}

// DataGetRes defines the response for querying dictionary data detail.
type DataGetRes struct {
	DictDataItem
}
