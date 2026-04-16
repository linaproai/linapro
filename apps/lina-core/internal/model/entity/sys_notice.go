// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysNotice is the golang structure for table sys_notice.
type SysNotice struct {
	Id        int64       `json:"id"        orm:"id"         description:"公告ID"`
	Title     string      `json:"title"     orm:"title"      description:"公告标题"`
	Type      int         `json:"type"      orm:"type"       description:"公告类型（1通知 2公告）"`
	Content   string      `json:"content"   orm:"content"    description:"公告内容"`
	FileIds   string      `json:"fileIds"   orm:"file_ids"   description:"附件文件ID列表，逗号分隔"`
	Status    int         `json:"status"    orm:"status"     description:"公告状态（0草稿 1已发布）"`
	Remark    string      `json:"remark"    orm:"remark"     description:"备注"`
	CreatedBy int64       `json:"createdBy" orm:"created_by" description:"创建者"`
	UpdatedBy int64       `json:"updatedBy" orm:"updated_by" description:"更新者"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"删除时间"`
}
