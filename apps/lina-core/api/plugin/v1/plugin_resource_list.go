package v1

import "github.com/gogf/gf/v2/frame/g"

// ResourceListReq is the request for querying plugin-owned backend resources.
type ResourceListReq struct {
	g.Meta   `path:"/plugins/{id}/resources/{resource}" method:"get" tags:"插件管理" summary:"查询插件资源数据" permission:"plugin:query" dc:"按插件通用资源契约查询插件自有的后端资源数据，资源的字段、过滤条件与排序规则均由插件目录下的后端实现注册定义"`
	Id       string `json:"id" v:"required|length:1,64" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Resource string `json:"resource" v:"required|length:1,64" dc:"插件资源标识，由插件自身在插件目录后端实现中注册" eg:"records"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码，从1开始" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页记录数，最大100" eg:"10"`
}

// ResourceListRes is the response for querying plugin resources.
type ResourceListRes struct {
	List  []map[string]interface{} `json:"list" dc:"插件资源记录列表，具体字段结构由插件自身资源声明决定" eg:"[]"`
	Total int                      `json:"total" dc:"总记录数" eg:"1"`
}
