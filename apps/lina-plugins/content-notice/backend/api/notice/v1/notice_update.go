// This file declares the update-notice request/response DTOs used by the
// content-notice source plugin.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Notice Update API

// UpdateReq defines the request for updating a notice.
type UpdateReq struct {
	g.Meta  `path:"/notice/{id}" method:"put" tags:"Notices" summary:"Update notification or announcement" dc:"Update the specified notification or announcement. All fields are optional." permission:"system:notice:edit"`
	Id      int64   `json:"id" v:"required" dc:"Announcement ID" eg:"1"`
	Title   *string `json:"title" dc:"Announcement title" eg:"System maintenance notification (update)"`
	Type    *int    `json:"type" dc:"Announcement type: 1=Notice 2=Announcement" eg:"1"`
	Content *string `json:"content" dc:"Announcement content (supports rich text HTML)" eg:"<p>Updated content</p>"`
	FileIds *string `json:"fileIds" dc:"Attachment file ID list, comma-separated" eg:"1,2,3"`
	Status  *int    `json:"status" dc:"Announcement status: 0=Draft 1=Published" eg:"1"`
	Remark  *string `json:"remark" dc:"Remark" eg:"Updated"`
}

// UpdateRes Notice update response
type UpdateRes struct{}
