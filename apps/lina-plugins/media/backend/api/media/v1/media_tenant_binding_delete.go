// This file declares media tenant strategy binding delete DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteTenantBindingReq defines the request for deleting one tenant strategy binding.
type DeleteTenantBindingReq struct {
	g.Meta   `path:"/media/tenant-bindings/{tenantId}" method:"delete" tags:"媒体管理" summary:"删除租户策略绑定" dc:"按租户ID删除租户策略绑定。" permission:"media:management:remove"`
	TenantId string `json:"tenantId" v:"required|length:1,255#租户ID不能为空|租户ID长度不能超过255个字符" dc:"租户ID" eg:"tenant-a"`
}

// DeleteTenantBindingRes defines the response for deleting one tenant strategy binding.
type DeleteTenantBindingRes struct {
	TenantId string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
}
