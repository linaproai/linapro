// This file defines DTOs for the runtime i18n message-bundle API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// RuntimeMessagesReq requests the aggregated runtime translation bundle for one locale.
type RuntimeMessagesReq struct {
	g.Meta `path:"/i18n/runtime/messages" method:"get" tags:"国际化" summary:"获取运行时国际化消息包" dc:"返回宿主与已接入资源聚合后的运行时国际化消息包，供登录页、管理工作台和宿主嵌入式插件页面加载"`
	Lang   string `json:"lang" dc:"目标语言编码；不传时按请求上下文自动解析，如 zh-CN、en-US" eg:"en-US"`
}

// RuntimeMessagesRes returns one runtime translation bundle.
type RuntimeMessagesRes struct {
	Locale   string                 `json:"locale" dc:"本次请求最终生效的语言编码" eg:"en-US"`
	Messages map[string]interface{} `json:"messages" dc:"聚合后的运行时国际化消息集合" eg:"{}"`
}
