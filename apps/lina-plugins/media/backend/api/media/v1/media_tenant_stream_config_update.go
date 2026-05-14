// This file declares media tenant stream config update DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateTenantStreamConfigReq defines the request for updating one tenant stream config.
type UpdateTenantStreamConfigReq struct {
	g.Meta        `path:"/media/tenant-stream-configs/{oldTenantId}" method:"put" tags:"媒体管理" summary:"修改租户流配置" dc:"按原租户ID修改租户流配置。" permission:"media:management:edit"`
	OldTenantId   string `json:"oldTenantId" v:"required|length:1,64#原租户ID不能为空|原租户ID长度不能超过64个字符" dc:"原租户ID" eg:"tenant-a"`
	TenantId      string `json:"tenantId" v:"required|length:1,64#租户ID不能为空|租户ID长度不能超过64个字符" dc:"租户ID" eg:"tenant-b"`
	MaxConcurrent int    `json:"maxConcurrent" v:"min:0#最大并发数不能小于0" dc:"最大并发数" eg:"100"`
	NodeNum       int    `json:"nodeNum" v:"min:0|max:255#节点编号不能小于0|节点编号不能大于255" dc:"节点编号" eg:"1"`
	Enable        int    `json:"enable" d:"1" dc:"1开启，0关闭" eg:"1"`
}

// UpdateTenantStreamConfigRes defines the response for updating one tenant stream config.
type UpdateTenantStreamConfigRes struct {
	TenantId string `json:"tenantId" dc:"租户ID" eg:"tenant-b"`
}
