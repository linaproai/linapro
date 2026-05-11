// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysFile is the golang structure for table sys_file.
type SysFile struct {
	Id        int64       `json:"id"        orm:"id"         description:"File ID"`
	TenantId  int         `json:"tenantId"  orm:"tenant_id"  description:"Owning tenant ID, 0 means PLATFORM"`
	Name      string      `json:"name"      orm:"name"       description:"Stored file name"`
	Original  string      `json:"original"  orm:"original"   description:"Original file name"`
	Suffix    string      `json:"suffix"    orm:"suffix"     description:"File suffix"`
	Scene     string      `json:"scene"     orm:"scene"      description:"Usage scene: avatar=user avatar, notice_image=notice image, notice_attachment=notice attachment, other=other"`
	Size      int64       `json:"size"      orm:"size"       description:"File size in bytes"`
	Hash      string      `json:"hash"      orm:"hash"       description:"File SHA-256 hash for deduplication"`
	Url       string      `json:"url"       orm:"url"        description:"File access URL"`
	Path      string      `json:"path"      orm:"path"       description:"File storage path"`
	Engine    string      `json:"engine"    orm:"engine"     description:"Storage engine: local=local storage"`
	CreatedBy int64       `json:"createdBy" orm:"created_by" description:"Uploader user ID"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}
