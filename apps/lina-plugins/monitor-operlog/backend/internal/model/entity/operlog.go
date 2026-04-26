// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Operlog is the golang structure for table operlog.
type Operlog struct {
	Id            int         `json:"id"            orm:"id"             description:"日志ID"`
	Title         string      `json:"title"         orm:"title"          description:"模块标题"`
	OperSummary   string      `json:"operSummary"   orm:"oper_summary"   description:"操作摘要"`
	RouteOwner    string      `json:"routeOwner"    orm:"route_owner"    description:"路由归属（core或插件ID）"`
	RouteMethod   string      `json:"routeMethod"   orm:"route_method"   description:"路由请求方式"`
	RoutePath     string      `json:"routePath"     orm:"route_path"     description:"路由路径"`
	RouteDocKey   string      `json:"routeDocKey"   orm:"route_doc_key"  description:"API文档结构化键"`
	OperType      string      `json:"operType"      orm:"oper_type"      description:"操作类型（create新增 update修改 delete删除 export导出 import导入 other其他）"`
	Method        string      `json:"method"        orm:"method"         description:"方法名称"`
	RequestMethod string      `json:"requestMethod" orm:"request_method" description:"请求方式（GET/POST/PUT/DELETE）"`
	OperName      string      `json:"operName"      orm:"oper_name"      description:"操作人员"`
	OperUrl       string      `json:"operUrl"       orm:"oper_url"       description:"请求URL"`
	OperIp        string      `json:"operIp"        orm:"oper_ip"        description:"操作IP地址"`
	OperParam     string      `json:"operParam"     orm:"oper_param"     description:"请求参数"`
	JsonResult    string      `json:"jsonResult"    orm:"json_result"    description:"返回参数"`
	Status        int         `json:"status"        orm:"status"         description:"操作状态（0成功 1失败）"`
	ErrorMsg      string      `json:"errorMsg"      orm:"error_msg"      description:"错误消息"`
	CostTime      int         `json:"costTime"      orm:"cost_time"      description:"耗时（毫秒）"`
	OperTime      *gtime.Time `json:"operTime"      orm:"oper_time"      description:"操作时间"`
}
