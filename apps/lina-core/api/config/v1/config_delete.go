package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Config Delete API

// DeleteReq defines the request for deleting a config.
type DeleteReq struct {
	g.Meta `path:"/config/{id}" method:"delete" tags:"Parameter Settings" summary:"Delete parameter settings" dc:"Soft delete the specified parameter setting record according to the parameter ID" permission:"system:config:remove"`
	Id     int `json:"id" v:"required" dc:"Parameter ID" eg:"1"`
}

// DeleteRes is the config delete response.
type DeleteRes struct{}
