// This file declares media tenant whitelist detail DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetTenantWhiteReq defines the request for querying one tenant whitelist entry.
type GetTenantWhiteReq struct {
	g.Meta   `path:"/media/tenant-whites/{tenantId}/{ip}" method:"get" tags:"媒体管理" summary:"查询租户白名单详情" dc:"按租户ID和白名单地址查询租户白名单详情。" permission:"media:management:query"`
	TenantId string `json:"tenantId" v:"required|length:1,64#租户ID不能为空|租户ID长度不能超过64个字符" dc:"租户ID" eg:"tenant-a"`
	Ip       string `json:"ip" v:"required|length:1,32#白名单地址不能为空|白名单地址长度不能超过32个字符" dc:"白名单地址" eg:"192.168.1.10"`
}

// GetTenantWhiteRes defines the response for querying one tenant whitelist entry.
type GetTenantWhiteRes = TenantWhiteListItem
