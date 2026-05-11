// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysKvCache is the golang structure for table sys_kv_cache.
type SysKvCache struct {
	Id         int64       `json:"id"         orm:"id"          description:"Primary key ID"`
	TenantId   int         `json:"tenantId"   orm:"tenant_id"   description:"Owning tenant ID, 0 means PLATFORM"`
	OwnerType  string      `json:"ownerType"  orm:"owner_type"  description:"Owner type: plugin=dynamic plugin, module=host module"`
	OwnerKey   string      `json:"ownerKey"   orm:"owner_key"   description:"Owner key: plugin ID or module name"`
	Namespace  string      `json:"namespace"  orm:"namespace"   description:"Cache namespace mapped to the host-cache resource identifier"`
	CacheKey   string      `json:"cacheKey"   orm:"cache_key"   description:"Cache key"`
	ValueKind  int         `json:"valueKind"  orm:"value_kind"  description:"Value type: 1=string, 2=integer"`
	ValueBytes string      `json:"valueBytes" orm:"value_bytes" description:"Cache byte value used by get/set"`
	ValueInt   int64       `json:"valueInt"   orm:"value_int"   description:"Cache integer value used by incr"`
	ExpireAt   *gtime.Time `json:"expireAt"   orm:"expire_at"   description:"Expiration time, NULL means never expires"`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:"Creation time"`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:"Update time"`
}
