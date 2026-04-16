package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeleteReq defines the request for deleting a user.
type DeleteReq struct {
	g.Meta `path:"/user/{id}" method:"delete" tags:"用户管理" summary:"删除用户" dc:"根据用户ID删除指定用户，不允许删除管理员账号" permission:"system:user:remove"`
	Id     int `json:"id" v:"required" dc:"用户ID" eg:"1"`
}

// DeleteRes defines the response for deleting a user.
type DeleteRes struct{}
