// This file defines DTOs for runtime i18n source diagnostics.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DiagnoseMessagesReq requests source diagnostics for one locale.
type DiagnoseMessagesReq struct {
	g.Meta    `path:"/i18n/messages/diagnostics" method:"get" tags:"国际化" summary:"诊断翻译来源" dc:"诊断目标语言每个翻译键的实际命中来源、是否回退到默认语言以及最终生效的语言，便于排查覆写优先级与缺失翻译问题" permission:"system:i18n:diagnose"`
	Locale    string `json:"locale" dc:"目标语言编码，不传时按请求上下文自动解析，如 zh-CN、en-US" eg:"en-US"`
	KeyPrefix string `json:"keyPrefix" dc:"按翻译键前缀过滤，如 menu.、plugin.demo.；不传则诊断全部" eg:"menu."`
}

// MessageDiagnosticItem describes the effective resolution result for one translation key.
type MessageDiagnosticItem struct {
	Key             string `json:"key" dc:"翻译键" eg:"menu.dashboard.title"`
	Value           string `json:"value" dc:"当前最终生效的翻译值" eg:"Workbench"`
	RequestedLocale string `json:"requestedLocale" dc:"请求的目标语言编码" eg:"en-US"`
	EffectiveLocale string `json:"effectiveLocale" dc:"实际提供该翻译值的语言编码" eg:"en-US"`
	FromFallback    bool   `json:"fromFallback" dc:"是否回退到默认语言：true=是 false=否" eg:"false"`
	SourceType      string `json:"sourceType" dc:"命中来源类型：host_file=宿主文件 plugin_file=插件文件 database=数据库覆写" eg:"database"`
	SourceKey       string `json:"sourceKey" dc:"命中来源标识，如 core、plugin-id 或具体作用域键" eg:"core"`
}

// DiagnoseMessagesRes returns runtime source diagnostics for one locale.
type DiagnoseMessagesRes struct {
	Locale        string                  `json:"locale" dc:"目标语言编码" eg:"en-US"`
	DefaultLocale string                  `json:"defaultLocale" dc:"当前宿主默认语言编码" eg:"zh-CN"`
	Total         int                     `json:"total" dc:"诊断返回的翻译键数量" eg:"128"`
	Items         []MessageDiagnosticItem `json:"items" dc:"翻译来源诊断结果列表" eg:"[]"`
}
