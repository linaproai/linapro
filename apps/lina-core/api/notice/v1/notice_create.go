package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Notice Create API

// CreateReq defines the request for creating a notice.
type CreateReq struct {
	g.Meta  `path:"/notice" method:"post" tags:"通知公告" summary:"创建通知公告" dc:"创建一条通知或公告，支持设置为草稿或直接发布，可附带附件文件" permission:"system:notice:add"`
	Title   string `json:"title" v:"required#请输入公告标题" dc:"公告标题" eg:"系统维护通知"`
	Type    int    `json:"type" v:"required|in:1,2#请选择公告类型|公告类型不正确" dc:"公告类型：1=通知 2=公告" eg:"1"`
	Content string `json:"content" v:"required#请输入公告内容" dc:"公告内容（支持富文本HTML）" eg:"<p>系统将于今晚进行维护升级</p>"`
	FileIds string `json:"fileIds" dc:"附件文件ID列表，逗号分隔，通过文件上传接口获取" eg:"1,2,3"`
	Status  *int   `json:"status" d:"0" dc:"公告状态：0=草稿 1=已发布" eg:"1"`
	Remark  string `json:"remark" dc:"备注" eg:"紧急通知"`
}

// CreateRes Notice create response
type CreateRes struct {
	Id int64 `json:"id" dc:"公告ID" eg:"1"`
}
