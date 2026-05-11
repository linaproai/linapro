// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Notice is the golang structure for table notice.
type Notice struct {
	Id        int64       `json:"id"        orm:"id"         description:"Notice ID"`
	TenantId  int         `json:"tenantId"  orm:"tenant_id"  description:"Owning tenant ID, 0 means PLATFORM"`
	Title     string      `json:"title"     orm:"title"      description:"Notice title"`
	Type      int         `json:"type"      orm:"type"       description:"Notice type: 1=notification, 2=announcement"`
	Content   string      `json:"content"   orm:"content"    description:"Notice content"`
	FileIds   string      `json:"fileIds"   orm:"file_ids"   description:"Attachment file ID list, comma-separated"`
	Status    int         `json:"status"    orm:"status"     description:"Notice status: 0=draft, 1=published"`
	Remark    string      `json:"remark"    orm:"remark"     description:"Remark"`
	CreatedBy int64       `json:"createdBy" orm:"created_by" description:"Creator"`
	UpdatedBy int64       `json:"updatedBy" orm:"updated_by" description:"Updater"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}
