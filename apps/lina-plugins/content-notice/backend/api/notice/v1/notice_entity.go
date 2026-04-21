// This file declares the plugin-local notice entity exposed in API responses
// so DTOs do not need to depend on host-internal entity packages.

package v1

import "github.com/gogf/gf/v2/os/gtime"

// NoticeEntity mirrors the plugin_content_notice table shape returned through the notice API.
type NoticeEntity struct {
	// Id is the notice primary key.
	Id int64 `json:"id" dc:"公告ID" eg:"1"`
	// Title is the notice title.
	Title string `json:"title" dc:"公告标题" eg:"系统维护通知"`
	// Type is the notice type (1=Notice 2=Announcement).
	Type int `json:"type" dc:"公告类型：1=通知 2=公告" eg:"1"`
	// Content is the notice body (may contain HTML rich text).
	Content string `json:"content" dc:"公告内容，支持富文本 HTML" eg:"<p>系统将于今晚进行维护升级</p>"`
	// FileIds is the comma-separated list of attachment file IDs.
	FileIds string `json:"fileIds" dc:"附件文件ID列表，多个以逗号分隔" eg:"1,2,3"`
	// Status is the notice publication status (0=Draft 1=Published).
	Status int `json:"status" dc:"公告状态：0=草稿 1=已发布" eg:"1"`
	// Remark is the optional free-form remark on the notice.
	Remark string `json:"remark" dc:"备注" eg:"紧急通知"`
	// CreatedBy is the creator user ID.
	CreatedBy int64 `json:"createdBy" dc:"创建者用户ID" eg:"1"`
	// UpdatedBy is the last-updater user ID.
	UpdatedBy int64 `json:"updatedBy" dc:"最后更新者用户ID" eg:"1"`
	// CreatedAt is the creation timestamp.
	CreatedAt *gtime.Time `json:"createdAt" dc:"创建时间" eg:"2026-04-21 10:00:00"`
	// UpdatedAt is the last-update timestamp.
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"最后更新时间" eg:"2026-04-21 10:30:00"`
	// DeletedAt is the soft-delete timestamp.
	DeletedAt *gtime.Time `json:"deletedAt" dc:"软删除时间，未删除时为空" eg:"null"`
}
