// This file declares media tenant stream config delete DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteTenantStreamConfigReq defines the request for deleting one tenant stream config.
type DeleteTenantStreamConfigReq struct {
	g.Meta   `path:"/media/tenant-stream-configs/{tenantId}" method:"delete" tags:"媒体管理" summary:"删除租户流配置" dc:"按租户ID删除租户流配置。" permission:"media:management:remove"`
	TenantId string `json:"tenantId" v:"required|length:1,64#租户ID不能为空|租户ID长度不能超过64个字符" dc:"租户ID" eg:"tenant-a"`
}

// DeleteTenantStreamConfigRes defines the response for deleting one tenant stream config.
type DeleteTenantStreamConfigRes struct {
	TenantId string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
}
