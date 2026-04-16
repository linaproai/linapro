package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Notice Update API

// UpdateReq defines the request for updating a notice.
type UpdateReq struct {
	g.Meta  `path:"/notice/{id}" method:"put" tags:"通知公告" summary:"更新通知公告" dc:"更新指定通知公告的信息，所有字段均为可选更新" permission:"system:notice:edit"`
	Id      int64   `json:"id" v:"required" dc:"公告ID" eg:"1"`
	Title   *string `json:"title" dc:"公告标题" eg:"系统维护通知（更新）"`
	Type    *int    `json:"type" dc:"公告类型：1=通知 2=公告" eg:"1"`
	Content *string `json:"content" dc:"公告内容（支持富文本HTML）" eg:"<p>更新后的内容</p>"`
	FileIds *string `json:"fileIds" dc:"附件文件ID列表，逗号分隔" eg:"1,2,3"`
	Status  *int    `json:"status" dc:"公告状态：0=草稿 1=已发布" eg:"1"`
	Remark  *string `json:"remark" dc:"备注" eg:"已更新"`
}

// UpdateRes Notice update response
type UpdateRes struct{}
