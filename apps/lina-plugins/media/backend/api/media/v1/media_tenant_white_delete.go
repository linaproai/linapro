// This file declares media tenant whitelist delete DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteTenantWhiteReq defines the request for deleting one tenant whitelist entry.
type DeleteTenantWhiteReq struct {
	g.Meta   `path:"/media/tenant-whites/{tenantId}/{ip}" method:"delete" tags:"媒体管理" summary:"删除租户白名单" dc:"按租户ID和白名单地址删除租户白名单。" permission:"media:management:remove"`
	TenantId string `json:"tenantId" v:"required|length:1,64#租户ID不能为空|租户ID长度不能超过64个字符" dc:"租户ID" eg:"tenant-a"`
	Ip       string `json:"ip" v:"required|length:1,32#白名单地址不能为空|白名单地址长度不能超过32个字符" dc:"白名单地址" eg:"192.168.1.10"`
}

// DeleteTenantWhiteRes defines the response for deleting one tenant whitelist entry.
type DeleteTenantWhiteRes struct {
	TenantId string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
	Ip       string `json:"ip" dc:"白名单地址" eg:"192.168.1.10"`
}
