package v1

import "github.com/gogf/gf/v2/frame/g"

// GetReq defines the request for querying post detail.
type GetReq struct {
	g.Meta `path:"/post/{id}" method:"get" tags:"Position Management" summary:"Get position details" dc:"Get detailed information about the position based on the position ID" permission:"system:post:query"`
	Id     int `json:"id" v:"required" dc:"Position ID" eg:"1"`
}

// GetRes is the response for post detail.
type GetRes struct {
	*PostEntity `dc:"Position information" eg:""`
}
