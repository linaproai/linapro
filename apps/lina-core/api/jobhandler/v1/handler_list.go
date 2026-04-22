package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq defines the request for querying registered job handlers.
type ListReq struct {
	g.Meta  `path:"/job/handler" method:"get" tags:"任务调度/插件处理器" summary:"获取处理器列表" dc:"查询当前宿主与插件已注册的任务处理器定义，供任务表单下拉选择" permission:"system:job:list"`
	Source  string `json:"source" dc:"按处理器来源筛选：host=宿主 plugin=插件，不传则查询全部" eg:"host"`
	Keyword string `json:"keyword" dc:"按处理器引用或展示名称关键字筛选" eg:"cleanup"`
}

// ListItem represents one registered job handler.
type ListItem struct {
	Ref         string `json:"ref" dc:"处理器唯一引用" eg:"host:cleanup-job-logs"`
	DisplayName string `json:"displayName" dc:"处理器展示名称" eg:"任务日志清理"`
	Description string `json:"description" dc:"处理器描述" eg:"按策略清理任务执行日志"`
	Source      string `json:"source" dc:"处理器来源：host=宿主 plugin=插件" eg:"host"`
	PluginId    string `json:"pluginId" dc:"来源插件ID；宿主处理器为空字符串" eg:"plugin-demo-source"`
}

// ListRes defines the response for querying registered job handlers.
type ListRes struct {
	List []*ListItem `json:"list" dc:"处理器列表" eg:"[]"`
}
