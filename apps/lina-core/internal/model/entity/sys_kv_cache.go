// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysKvCache is the golang structure for table sys_kv_cache.
type SysKvCache struct {
	Id         int64       `json:"id"         orm:"id"          description:"主键ID"`
	OwnerType  string      `json:"ownerType"  orm:"owner_type"  description:"所属类型：plugin=动态插件 module=宿主模块"`
	OwnerKey   string      `json:"ownerKey"   orm:"owner_key"   description:"所属标识：插件ID或模块名"`
	Namespace  string      `json:"namespace"  orm:"namespace"   description:"缓存命名空间，对应 host-cache 资源标识"`
	CacheKey   string      `json:"cacheKey"   orm:"cache_key"   description:"缓存键"`
	ValueKind  int         `json:"valueKind"  orm:"value_kind"  description:"值类型：1=字符串 2=整数"`
	ValueBytes []byte      `json:"valueBytes" orm:"value_bytes" description:"缓存字节值，供 get/set 使用"`
	ValueInt   int64       `json:"valueInt"   orm:"value_int"   description:"缓存整数值，供 incr 使用"`
	ExpireAt   *gtime.Time `json:"expireAt"   orm:"expire_at"   description:"过期时间，NULL表示永不过期"`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:"创建时间"`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:"更新时间"`
}
