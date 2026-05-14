package v1

import "github.com/gogf/gf/v2/frame/g"

// Config Get API

// GetReq defines the request for getting config detail by ID.
type GetReq struct {
	g.Meta `path:"/config/{id}" method:"get" tags:"Parameter Settings" summary:"Get parameter setting details" dc:"Get detailed information about parameter settings based on parameter ID" permission:"system:config:query"`
	Id     int `json:"id" v:"required" dc:"Parameter ID" eg:"1"`
}

// GetRes is the config detail response.
type GetRes struct {
	ConfigItem
}
