// This file declares media tenant whitelist create DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateTenantWhiteReq defines the request for creating one tenant whitelist entry.
type CreateTenantWhiteReq struct {
	g.Meta      `path:"/media/tenant-whites" method:"post" tags:"媒体管理" summary:"新增租户白名单" dc:"新增一条租户白名单地址配置。" permission:"media:management:add"`
	TenantId    string `json:"tenantId" v:"required|length:1,64#租户ID不能为空|租户ID长度不能超过64个字符" dc:"租户ID" eg:"tenant-a"`
	Ip          string `json:"ip" v:"required|length:1,32#白名单地址不能为空|白名单地址长度不能超过32个字符" dc:"白名单地址" eg:"192.168.1.10"`
	Description string `json:"description" v:"max-length:32#白名单描述长度不能超过32个字符" dc:"白名单描述" eg:"总部出口"`
	Enable      int    `json:"enable" d:"1" dc:"1开启，0关闭" eg:"1"`
}

// CreateTenantWhiteRes defines the response for creating one tenant whitelist entry.
type CreateTenantWhiteRes struct {
	TenantId string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
	Ip       string `json:"ip" dc:"白名单地址" eg:"192.168.1.10"`
}
