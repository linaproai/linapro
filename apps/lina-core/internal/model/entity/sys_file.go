// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysFile is the golang structure for table sys_file.
type SysFile struct {
	Id        int64       `json:"id"        orm:"id"         description:"文件ID"`
	Name      string      `json:"name"      orm:"name"       description:"存储文件名"`
	Original  string      `json:"original"  orm:"original"   description:"原始文件名"`
	Suffix    string      `json:"suffix"    orm:"suffix"     description:"文件后缀"`
	Scene     string      `json:"scene"     orm:"scene"      description:"使用场景"`
	Size      int64       `json:"size"      orm:"size"       description:"文件大小（字节）"`
	Hash      string      `json:"hash"      orm:"hash"       description:"文件SHA-256散列值，用于去重"`
	Url       string      `json:"url"       orm:"url"        description:"文件访问URL"`
	Path      string      `json:"path"      orm:"path"       description:"文件存储路径"`
	Engine    string      `json:"engine"    orm:"engine"     description:"存储引擎：local=本地"`
	CreatedBy int64       `json:"createdBy" orm:"created_by" description:"上传者用户ID"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"删除时间"`
}
