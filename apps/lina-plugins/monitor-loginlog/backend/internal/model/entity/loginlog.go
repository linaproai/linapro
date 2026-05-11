// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Loginlog is the golang structure for table loginlog.
type Loginlog struct {
	Id                 int         `json:"id"                 orm:"id"                     description:"Log ID"`
	TenantId           int         `json:"tenantId"           orm:"tenant_id"              description:"Owning tenant ID, 0 means PLATFORM"`
	ActingUserId       int         `json:"actingUserId"       orm:"acting_user_id"         description:"Actual acting user ID for platform operations or impersonation"`
	OnBehalfOfTenantId int         `json:"onBehalfOfTenantId" orm:"on_behalf_of_tenant_id" description:"Target tenant ID when a platform administrator acts on behalf of a tenant"`
	IsImpersonation    bool        `json:"isImpersonation"    orm:"is_impersonation"       description:"Whether this log was produced during tenant impersonation"`
	UserName           string      `json:"userName"           orm:"user_name"              description:"Login account"`
	Status             int         `json:"status"             orm:"status"                 description:"Login status: 0=succeeded, 1=failed"`
	Ip                 string      `json:"ip"                 orm:"ip"                     description:"Login IP address"`
	Browser            string      `json:"browser"            orm:"browser"                description:"Browser type"`
	Os                 string      `json:"os"                 orm:"os"                     description:"Operating system"`
	Msg                string      `json:"msg"                orm:"msg"                    description:"Prompt message"`
	LoginTime          *gtime.Time `json:"loginTime"          orm:"login_time"             description:"Login time"`
}
