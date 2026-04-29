// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Operlog is the golang structure for table operlog.
type Operlog struct {
	Id            int         `json:"id"            orm:"id"             description:"Log ID"`
	Title         string      `json:"title"         orm:"title"          description:"Module title"`
	OperSummary   string      `json:"operSummary"   orm:"oper_summary"   description:"Operation summary"`
	RouteOwner    string      `json:"routeOwner"    orm:"route_owner"    description:"Route owner: core or plugin ID"`
	RouteMethod   string      `json:"routeMethod"   orm:"route_method"   description:"Route request method"`
	RoutePath     string      `json:"routePath"     orm:"route_path"     description:"Route path"`
	RouteDocKey   string      `json:"routeDocKey"   orm:"route_doc_key"  description:"API documentation structured key"`
	OperType      string      `json:"operType"      orm:"oper_type"      description:"Operation type: create=create, update=update, delete=delete, export=export, import=import, other=other"`
	Method        string      `json:"method"        orm:"method"         description:"Method name"`
	RequestMethod string      `json:"requestMethod" orm:"request_method" description:"Request method: GET/POST/PUT/DELETE"`
	OperName      string      `json:"operName"      orm:"oper_name"      description:"Operator"`
	OperUrl       string      `json:"operUrl"       orm:"oper_url"       description:"Request URL"`
	OperIp        string      `json:"operIp"        orm:"oper_ip"        description:"Operation IP address"`
	OperParam     string      `json:"operParam"     orm:"oper_param"     description:"Request parameters"`
	JsonResult    string      `json:"jsonResult"    orm:"json_result"    description:"Response parameters"`
	Status        int         `json:"status"        orm:"status"         description:"Operation status: 0=succeeded, 1=failed"`
	ErrorMsg      string      `json:"errorMsg"      orm:"error_msg"      description:"Error message"`
	CostTime      int         `json:"costTime"      orm:"cost_time"      description:"Duration in milliseconds"`
	OperTime      *gtime.Time `json:"operTime"      orm:"oper_time"      description:"Operation time"`
}
