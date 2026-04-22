// This file defines the backend-summary route DTOs for the dynamic plugin
// sample.

package v1

import "github.com/gogf/gf/v2/util/gmeta"

// BackendSummaryReq is the request for querying the dynamic plugin backend execution summary.
type BackendSummaryReq struct {
	gmeta.Meta `path:"/backend-summary" method:"get" tags:"动态插件示例" summary:"查询动态插件后端执行摘要" dc:"通过宿主固定前缀 /api/v1/extensions/{pluginId}/... 分发到 plugin-demo-dynamic 的 Wasm bridge 运行时，返回动态插件当前桥接执行摘要，包含插件标识、路由信息、当前登录用户等上下文字段，用于演示动态插件后端路由的完整工作流" access:"login" permission:"plugin-demo-dynamic:backend:view" operLog:"other"`
}

// BackendSummaryRes is the response for querying the dynamic plugin backend execution summary.
type BackendSummaryRes struct {
	Message       string  `json:"message" dc:"动态插件后端执行说明，描述当前请求经由 Wasm bridge 运行时处理的路径和方式" eg:"This backend example is executed through the plugin-demo-dynamic Wasm bridge runtime."`
	PluginID      string  `json:"pluginId" dc:"当前执行该请求的动态插件唯一标识" eg:"plugin-demo-dynamic"`
	PublicPath    string  `json:"publicPath" dc:"当前命中的宿主公开路由路径" eg:"/api/v1/extensions/plugin-demo-dynamic/backend-summary"`
	Access        string  `json:"access" dc:"当前动态路由的访问级别：login=需要登录 public=匿名可访问" eg:"login"`
	Permission    string  `json:"permission" dc:"当前动态路由的权限标识；匿名路由时为空字符串" eg:"plugin-demo-dynamic:backend:view"`
	Authenticated bool    `json:"authenticated" dc:"当前请求是否带有宿主认证身份：true=已认证 false=匿名" eg:"true"`
	Username      *string `json:"username,omitempty" dc:"当前登录用户名；匿名请求时为空" eg:"admin"`
	IsSuperAdmin  *bool   `json:"isSuperAdmin,omitempty" dc:"当前身份是否为超级管理员；匿名请求时为空" eg:"true"`
}
