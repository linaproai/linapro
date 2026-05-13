// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaStreamAlias is the golang structure of table media_stream_alias for DAO operations like Where/Data.
type MediaStreamAlias struct {
	g.Meta     `orm:"table:media_stream_alias, do:true"`
	Id         any         // ID
	Alias      any         // 流别名
	AutoRemove any         // 是否自动移除：1是，0否
	StreamPath any         // 真实流路径
	CreateTime *gtime.Time // 创建时间
}
