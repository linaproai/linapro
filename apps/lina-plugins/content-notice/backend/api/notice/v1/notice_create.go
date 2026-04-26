// This file declares the create-notice request/response DTOs used by the
// content-notice source plugin.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Notice Create API

// CreateReq defines the request for creating a notice.
type CreateReq struct {
	g.Meta  `path:"/notice" method:"post" tags:"Notices" summary:"Create notification or announcement" dc:"Create a notification or announcement as either a draft or a published item, with optional attachments." permission:"system:notice:add"`
	Title   string `json:"title" v:"required#gf.gvalid.rule.required" dc:"Announcement title" eg:"System maintenance notification"`
	Type    int    `json:"type" v:"required|in:1,2#gf.gvalid.rule.required|gf.gvalid.rule.in" dc:"Announcement type: 1=Notice 2=Announcement" eg:"1"`
	Content string `json:"content" v:"required#gf.gvalid.rule.required" dc:"Announcement content (supports rich text HTML)" eg:"<p>The system will be undergoing maintenance and upgrade tonight</p>"`
	FileIds string `json:"fileIds" dc:"Attachment file ID list, comma-separated and geted from the file upload API" eg:"1,2,3"`
	Status  *int   `json:"status" d:"0" dc:"Announcement status: 0=Draft 1=Published" eg:"1"`
	Remark  string `json:"remark" dc:"Remark" eg:"Emergency notification"`
}

// CreateRes Notice create response
type CreateRes struct {
	Id int64 `json:"id" dc:"Announcement ID" eg:"1"`
}
