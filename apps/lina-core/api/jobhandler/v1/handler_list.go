package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq defines the request for querying registered job handlers.
type ListReq struct {
	g.Meta  `path:"/job/handler" method:"get" tags:"Job Scheduling / Plugin Handlers" summary:"Get processor list" dc:"Query the registered task processor definitions of the current host and plugin for dropdown selection in the task form." permission:"system:job:list"`
	Source  string `json:"source" dc:"Filter by processor source: host=host plugin=plugin, if not passed, query all" eg:"host"`
	Keyword string `json:"keyword" dc:"Filter by processor reference or display name keyword" eg:"cleanup"`
}

// ListItem represents one registered job handler.
type ListItem struct {
	Ref         string `json:"ref" dc:"Processor unique reference" eg:"host:cleanup-job-logs"`
	DisplayName string `json:"displayName" dc:"Processor display name" eg:"Task log cleaning"`
	Description string `json:"description" dc:"Processor description" eg:"Clean task execution logs according to policy"`
	Source      string `json:"source" dc:"Processor source: host=host plugin=plugin" eg:"host"`
	PluginId    string `json:"pluginId" dc:"Source plugin ID; host processor is an empty string" eg:"plugin-demo-source"`
}

// ListRes defines the response for querying registered job handlers.
type ListRes struct {
	List []*ListItem `json:"list" dc:"Processor list" eg:"[]"`
}
