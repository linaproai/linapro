// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysKvCache is the golang structure of table sys_kv_cache for DAO operations like Where/Data.
type SysKvCache struct {
	g.Meta     `orm:"table:sys_kv_cache, do:true"`
	Id         any         // 主键ID
	OwnerType  any         // 所属类型：plugin=动态插件 module=宿主模块
	OwnerKey   any         // 所属标识：插件ID或模块名
	Namespace  any         // 缓存命名空间，对应 host-cache 资源标识
	CacheKey   any         // 缓存键
	ValueKind  any         // 值类型：1=字符串 2=整数
	ValueBytes []byte      // 缓存字节值，供 get/set 使用
	ValueInt   any         // 缓存整数值，供 incr 使用
	ExpireAt   *gtime.Time // 过期时间，NULL表示永不过期
	CreatedAt  *gtime.Time // 创建时间
	UpdatedAt  *gtime.Time // 更新时间
}
