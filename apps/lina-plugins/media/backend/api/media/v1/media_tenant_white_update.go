// This file declares media tenant whitelist update DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateTenantWhiteReq defines the request for updating one tenant whitelist entry.
type UpdateTenantWhiteReq struct {
	g.Meta      `path:"/media/tenant-whites/{oldTenantId}/{oldIp}" method:"put" tags:"媒体管理" summary:"修改租户白名单" dc:"按原租户ID和原白名单地址修改租户白名单配置。" permission:"media:management:edit"`
	OldTenantId string `json:"oldTenantId" v:"required|length:1,64#原租户ID不能为空|原租户ID长度不能超过64个字符" dc:"原租户ID" eg:"tenant-a"`
	OldIp       string `json:"oldIp" v:"required|length:1,32#原白名单地址不能为空|原白名单地址长度不能超过32个字符" dc:"原白名单地址" eg:"192.168.1.10"`
	TenantId    string `json:"tenantId" v:"required|length:1,64#租户ID不能为空|租户ID长度不能超过64个字符" dc:"租户ID" eg:"tenant-b"`
	Ip          string `json:"ip" v:"required|length:1,32#白名单地址不能为空|白名单地址长度不能超过32个字符" dc:"白名单地址" eg:"192.168.1.11"`
	Description string `json:"description" v:"max-length:32#白名单描述长度不能超过32个字符" dc:"白名单描述" eg:"总部出口"`
	Enable      int    `json:"enable" d:"1" dc:"1开启，0关闭" eg:"1"`
}

// UpdateTenantWhiteRes defines the response for updating one tenant whitelist entry.
type UpdateTenantWhiteRes struct {
	TenantId string `json:"tenantId" dc:"租户ID" eg:"tenant-b"`
	Ip       string `json:"ip" dc:"白名单地址" eg:"192.168.1.11"`
}
