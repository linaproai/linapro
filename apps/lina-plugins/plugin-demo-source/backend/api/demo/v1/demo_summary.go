package v1

import "github.com/gogf/gf/v2/frame/g"

// SummaryReq is the request for querying plugin-demo-source summary.
type SummaryReq struct {
	g.Meta `path:"/plugins/plugin-demo-source/summary" method:"get" tags:"源码插件示例" summary:"查询源码插件示例摘要" dc:"返回 plugin-demo-source 页面展示所需的简要介绍文案，用于验证源码插件菜单页可读取插件后端接口数据" permission:"plugin-demo-source:example:view"`
}

// SummaryRes is the response for querying plugin-demo-source summary.
type SummaryRes struct {
	Message string `json:"message" dc:"页面展示使用的简要介绍文案，来自插件后端接口" eg:"这是一条来自 plugin-demo-source 接口的简要介绍，用于验证源码插件菜单页可读取插件后端数据。"`
}
