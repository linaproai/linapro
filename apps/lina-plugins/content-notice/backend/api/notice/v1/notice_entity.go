// This file declares the plugin-local notice entity exposed in API responses
// so DTOs do not need to depend on host-internal entity packages.

package v1

import "github.com/gogf/gf/v2/os/gtime"

// NoticeEntity mirrors the plugin_content_notice table shape returned through the notice API.
type NoticeEntity struct {
	// Id is the notice primary key.
	Id int64 `json:"id" dc:"Announcement ID" eg:"1"`
	// Title is the notice title.
	Title string `json:"title" dc:"Announcement title" eg:"System maintenance notification"`
	// Type is the notice type (1=Notice 2=Announcement).
	Type int `json:"type" dc:"Announcement type: 1=Notice 2=Announcement" eg:"1"`
	// Content is the notice body (may contain HTML rich text).
	Content string `json:"content" dc:"Announcement content, supports rich text HTML" eg:"<p>The system will be undergoing maintenance and upgrade tonight</p>"`
	// FileIds is the comma-separated list of attachment file IDs.
	FileIds string `json:"fileIds" dc:"Attachment file ID list, comma-separated" eg:"1,2,3"`
	// Status is the notice publication status (0=Draft 1=Published).
	Status int `json:"status" dc:"Announcement status: 0=Draft 1=Published" eg:"1"`
	// Remark is the optional free-form remark on the notice.
	Remark string `json:"remark" dc:"Remark" eg:"Emergency notification"`
	// CreatedBy is the creator user ID.
	CreatedBy int64 `json:"createdBy" dc:"Creator user ID" eg:"1"`
	// UpdatedBy is the last-updater user ID.
	UpdatedBy int64 `json:"updatedBy" dc:"Last updated user ID" eg:"1"`
	// CreatedAt is the Creation timestamp.
	CreatedAt *gtime.Time `json:"createdAt" dc:"Creation time" eg:"2026-04-21 10:00:00"`
	// UpdatedAt is the last-update timestamp.
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"Last updated time" eg:"2026-04-21 10:30:00"`
	// DeletedAt is the soft-delete timestamp.
	DeletedAt *gtime.Time `json:"deletedAt" dc:"Soft deletion time, empty if not deleted" eg:"null"`
}
