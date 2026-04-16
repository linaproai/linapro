package v1

import "github.com/gogf/gf/v2/frame/g"

// BackendSummaryReq is the request for querying the dynamic plugin backend execution summary.
type BackendSummaryReq struct {
	g.Meta `path:"/backend-summary" method:"get" tags:"动态插件示例" summary:"查询动态插件后端执行摘要" dc:"通过宿主固定前缀 /api/v1/extensions/{pluginId}/... 分发到 plugin-demo-dynamic 的 Wasm bridge 运行时，返回动态插件当前桥接执行摘要，包含插件标识、路由信息、当前登录用户等上下文字段，用于演示动态插件后端路由的完整工作流" access:"login" permission:"plugin-demo-dynamic:backend:view" operLog:"other"`
}

// BackendSummaryRes is the response for querying the dynamic plugin backend execution summary.
type BackendSummaryRes struct {
	Message string `json:"message" dc:"动态插件后端执行说明，描述当前请求经由 Wasm bridge 运行时处理的路径和方式" eg:"This backend example is executed through the plugin-demo-dynamic Wasm bridge runtime."`
}
