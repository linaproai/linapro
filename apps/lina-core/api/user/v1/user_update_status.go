package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UpdateStatusReq defines the request for updating a user's status.
type UpdateStatusReq struct {
	g.Meta `path:"/user/{id}/status" method:"put" tags:"用户管理" summary:"更新用户状态" dc:"启用或停用指定用户账号" permission:"system:user:edit"`
	Id     int `json:"id" v:"required" dc:"用户ID" eg:"1"`
	Status int `json:"status" v:"in:0,1#状态值无效" dc:"状态：1=正常 0=停用" eg:"1"`
}

// UpdateStatusRes defines the response for updating a user's status.
type UpdateStatusRes struct{}
