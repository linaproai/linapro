// This file declares the get-notice request/response DTOs used by the
// content-notice source plugin.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Notice Get API

// GetReq defines the request for retrieving notice details.
type GetReq struct {
	g.Meta `path:"/notice/{id}" method:"get" tags:"Notices" summary:"Get notification or announcement details" dc:"Get notification or announcement details by ID, including content and creator information." permission:"system:notice:query"`
	Id     int64 `json:"id" v:"required" dc:"Announcement ID" eg:"1"`
}

// GetRes Notice detail response
type GetRes struct {
	*NoticeEntity
	CreatedByName string `json:"createdByName" dc:"Creator username" eg:"admin"`
}
