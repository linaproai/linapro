package v1

import "github.com/gogf/gf/v2/frame/g"

// DetailReq defines the request for querying one registered job handler detail.
type DetailReq struct {
	g.Meta `path:"/job/handler/{ref}" method:"get" tags:"任务处理器注册表" summary:"获取处理器详情" dc:"根据处理器引用查询详情，返回描述信息和参数 Schema；包含斜杠的 ref 需进行 URL 编码" permission:"system:job:list"`
	Ref    string `json:"ref" v:"required" dc:"处理器唯一引用" eg:"host:cleanup-job-logs"`
}

// DetailRes defines the response for querying one registered job handler detail.
type DetailRes struct {
	Ref          string `json:"ref" dc:"处理器唯一引用" eg:"host:cleanup-job-logs"`
	DisplayName  string `json:"displayName" dc:"处理器展示名称" eg:"任务日志清理"`
	Description  string `json:"description" dc:"处理器描述" eg:"按策略清理任务执行日志"`
	Source       string `json:"source" dc:"处理器来源：host=宿主 plugin=插件" eg:"host"`
	PluginId     string `json:"pluginId" dc:"来源插件ID；宿主处理器为空字符串" eg:"plugin-demo-source"`
	ParamsSchema string `json:"paramsSchema" dc:"处理器参数 JSON Schema 文本" eg:"{\"type\":\"object\",\"properties\":{}}"`
}
