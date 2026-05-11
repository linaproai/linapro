package v1

import (
	"github.com/gogf/gf/v2/os/gtime"

	"lina-plugin-monitor-operlog/backend/internal/model/operlogtype"
)

// OperLogEntity represents one operation-log record returned by plugin APIs.
type OperLogEntity struct {
	Id                 int                  `json:"id" dc:"Log ID" eg:"1"`
	TenantId           int                  `json:"tenantId" dc:"Owning tenant ID, where 0 means platform" eg:"1001"`
	ActingUserId       int                  `json:"actingUserId" dc:"Actual acting user ID for platform operations or impersonation" eg:"1"`
	OnBehalfOfTenantId int                  `json:"onBehalfOfTenantId" dc:"Target tenant ID when a platform administrator acts on behalf of a tenant" eg:"1001"`
	IsImpersonation    bool                 `json:"isImpersonation" dc:"Whether this log was produced during tenant impersonation" eg:"true"`
	Title              string               `json:"title" dc:"Module title" eg:"User Management"`
	OperSummary        string               `json:"operSummary" dc:"Operation summary" eg:"Delete user"`
	OperType           operlogtype.OperType `json:"operType" dc:"Operation type: create=new update=modify delete=delete export=export import=import other=other" eg:"delete"`
	Method             string               `json:"method" dc:"Method name" eg:"/user/1"`
	RequestMethod      string               `json:"requestMethod" dc:"Request method" eg:"DELETE"`
	OperName           string               `json:"operName" dc:"Operator" eg:"admin"`
	OperUrl            string               `json:"operUrl" dc:"Request URL" eg:"/api/v1/user/1"`
	OperIp             string               `json:"operIp" dc:"Operation IP address" eg:"127.0.0.1"`
	OperParam          string               `json:"operParam" dc:"Request parameters" eg:"{"id":1}"`
	JsonResult         string               `json:"jsonResult" dc:"Return parameters" eg:"{"code":0}"`
	Status             int                  `json:"status" dc:"Operation status: 0=success 1=failure" eg:"0"`
	ErrorMsg           string               `json:"errorMsg" dc:"Error message" eg:""`
	CostTime           int                  `json:"costTime" dc:"Time taken (milliseconds)" eg:"32"`
	OperTime           *gtime.Time          `json:"operTime" dc:"Operating time" eg:"2025-01-01 12:00:00"`
}
