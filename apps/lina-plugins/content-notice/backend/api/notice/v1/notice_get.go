// This file declares the get-notice request/response DTOs used by the
// content-notice source plugin.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Notice Get API

// GetReq defines the request for retrieving notice details.
type GetReq struct {
	g.Meta `path:"/notice/{id}" method:"get" tags:"通知公告" summary:"获取通知公告详情" dc:"根据公告ID获取通知公告的详细信息，包括内容和创建者信息" permission:"system:notice:query"`
	Id     int64 `json:"id" v:"required" dc:"公告ID" eg:"1"`
}

// GetRes Notice detail response
type GetRes struct {
	*NoticeEntity
	CreatedByName string `json:"createdByName" dc:"创建者用户名" eg:"admin"`
}
