// This file declares media tenant stream config create DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateTenantStreamConfigReq defines the request for creating one tenant stream config.
type CreateTenantStreamConfigReq struct {
	g.Meta        `path:"/media/tenant-stream-configs" method:"post" tags:"媒体管理" summary:"新增租户流配置" dc:"新增一条租户流配置。" permission:"media:management:add"`
	TenantId      string `json:"tenantId" v:"required|length:1,64#租户ID不能为空|租户ID长度不能超过64个字符" dc:"租户ID" eg:"tenant-a"`
	MaxConcurrent int    `json:"maxConcurrent" v:"min:0#最大并发数不能小于0" dc:"最大并发数" eg:"100"`
	NodeNum       int    `json:"nodeNum" v:"min:0|max:255#节点编号不能小于0|节点编号不能大于255" dc:"节点编号" eg:"1"`
	Enable        int    `json:"enable" d:"1" dc:"1开启，0关闭" eg:"1"`
}

// CreateTenantStreamConfigRes defines the response for creating one tenant stream config.
type CreateTenantStreamConfigRes struct {
	TenantId string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
}
