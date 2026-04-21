package v1

import "github.com/gogf/gf/v2/os/gtime"

// OperLogEntity represents one operation-log record returned by plugin APIs.
type OperLogEntity struct {
	Id            int         `json:"id" dc:"日志ID" eg:"1"`
	Title         string      `json:"title" dc:"模块标题" eg:"用户管理"`
	OperSummary   string      `json:"operSummary" dc:"操作摘要" eg:"删除用户"`
	OperType      int         `json:"operType" dc:"操作类型：1=新增 2=修改 3=删除 4=导出 5=导入 6=其他" eg:"3"`
	Method        string      `json:"method" dc:"方法名称" eg:"/user/1"`
	RequestMethod string      `json:"requestMethod" dc:"请求方式" eg:"DELETE"`
	OperName      string      `json:"operName" dc:"操作人员" eg:"admin"`
	OperUrl       string      `json:"operUrl" dc:"请求URL" eg:"/api/v1/user/1"`
	OperIp        string      `json:"operIp" dc:"操作IP地址" eg:"127.0.0.1"`
	OperParam     string      `json:"operParam" dc:"请求参数" eg:"{"id":1}"`
	JsonResult    string      `json:"jsonResult" dc:"返回参数" eg:"{"code":0}"`
	Status        int         `json:"status" dc:"操作状态：0=成功 1=失败" eg:"0"`
	ErrorMsg      string      `json:"errorMsg" dc:"错误消息" eg:""`
	CostTime      int         `json:"costTime" dc:"耗时（毫秒）" eg:"32"`
	OperTime      *gtime.Time `json:"operTime" dc:"操作时间" eg:"2025-01-01 12:00:00"`
}
