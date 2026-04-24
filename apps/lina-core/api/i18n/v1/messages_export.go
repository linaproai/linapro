// This file defines DTOs for exporting flat i18n runtime messages.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ExportMessagesReq requests one flat runtime message export for the given locale.
type ExportMessagesReq struct {
	g.Meta `path:"/i18n/messages/export" method:"get" tags:"国际化" summary:"导出国际化消息" dc:"导出指定语言的扁平国际化消息集合，支持导出有效聚合结果或当前语言原始资源，用于交付维护、离线校对和再次导入" permission:"system:i18n:export"`
	Locale string `json:"locale" dc:"目标语言编码，不传时按请求上下文自动解析，如 zh-CN、en-US" eg:"en-US"`
	Raw    bool   `json:"raw" dc:"是否仅导出当前语言原始资源：true=仅导出当前语言聚合结果 false=导出带默认语言回退的有效结果" eg:"false"`
}

// ExportMessagesRes returns one flat runtime message export payload.
type ExportMessagesRes struct {
	Locale        string            `json:"locale" dc:"本次导出的目标语言编码" eg:"en-US"`
	DefaultLocale string            `json:"defaultLocale" dc:"当前宿主默认语言编码" eg:"zh-CN"`
	Mode          string            `json:"mode" dc:"导出模式：effective=带默认语言回退的有效结果 raw=当前语言原始资源" eg:"effective"`
	Total         int               `json:"total" dc:"导出的翻译键数量" eg:"128"`
	Messages      map[string]string `json:"messages" dc:"按扁平 key 输出的国际化消息集合" eg:"{}"`
}
