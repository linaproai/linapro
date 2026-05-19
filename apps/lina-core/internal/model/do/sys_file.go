// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysFile is the golang structure of table sys_file for DAO operations like Where/Data.
type SysFile struct {
	g.Meta    `orm:"table:sys_file, do:true"`
	Id        any        // File ID
	TenantId  any        // Owning tenant ID, 0 means PLATFORM
	Name      any        // Stored file name
	Original  any        // Original file name
	Suffix    any        // File suffix
	Scene     any        // Usage scene: avatar=user avatar, notice_image=notice image, notice_attachment=notice attachment, other=other
	Size      any        // File size in bytes
	Hash      any        // File SHA-256 hash for deduplication
	Url       any        // File access URL
	Path      any        // File storage path
	Engine    any        // Storage engine: local=local storage
	CreatedBy any        // Uploader user ID
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
	DeletedAt *time.Time // Deletion time
}
