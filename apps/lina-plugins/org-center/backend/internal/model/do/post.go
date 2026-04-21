// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Post is the golang structure of table plugin_org_center_post for DAO operations like Where/Data.
type Post struct {
	g.Meta    `orm:"table:plugin_org_center_post, do:true"`
	Id        any         // 岗位ID
	DeptId    any         // 所属部门ID
	Code      any         // 岗位编码
	Name      any         // 岗位名称
	Sort      any         // 显示排序
	Status    any         // 状态（0停用 1正常）
	Remark    any         // 备注
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
	DeletedAt *gtime.Time // 删除时间
}
