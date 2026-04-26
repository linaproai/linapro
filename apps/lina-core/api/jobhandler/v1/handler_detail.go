package v1

import "github.com/gogf/gf/v2/frame/g"

// DetailReq defines the request for querying one registered job handler detail.
type DetailReq struct {
	g.Meta `path:"/job/handler/{ref}" method:"get" tags:"Job Scheduling / Plugin Handlers" summary:"Get processor details" dc:"According to the processor reference query details, description information and parameter schema are returned; refs containing slashes need to be URL encoded" permission:"system:job:list"`
	Ref    string `json:"ref" v:"required" dc:"Processor unique reference" eg:"host:cleanup-job-logs"`
}

// DetailRes defines the response for querying one registered job handler detail.
type DetailRes struct {
	Ref          string `json:"ref" dc:"Processor unique reference" eg:"host:cleanup-job-logs"`
	DisplayName  string `json:"displayName" dc:"Processor display name" eg:"Task log cleaning"`
	Description  string `json:"description" dc:"Processor description" eg:"Clean task execution logs according to policy"`
	Source       string `json:"source" dc:"Processor source: host=host plugin=plugin" eg:"host"`
	PluginId     string `json:"pluginId" dc:"Source plugin ID; host processor is an empty string" eg:"plugin-demo-source"`
	ParamsSchema string `json:"paramsSchema" dc:"Processor parameters JSON Schema text" eg:"{\"type\":\"object\",\"properties\":{}}"`
}
