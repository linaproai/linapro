// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysFile is the golang structure of table sys_file for DAO operations like Where/Data.
type SysFile struct {
	g.Meta    `orm:"table:sys_file, do:true"`
	Id        any         // 文件ID
	Name      any         // 存储文件名
	Original  any         // 原始文件名
	Suffix    any         // 文件后缀
	Scene     any         // 使用场景
	Size      any         // 文件大小（字节）
	Hash      any         // 文件SHA-256散列值，用于去重
	Url       any         // 文件访问URL
	Path      any         // 文件存储路径
	Engine    any         // 存储引擎：local=本地
	CreatedBy any         // 上传者用户ID
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
	DeletedAt *gtime.Time // 删除时间
}
