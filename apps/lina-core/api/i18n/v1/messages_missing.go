// This file defines DTOs for checking missing i18n messages.

package v1

import "github.com/gogf/gf/v2/frame/g"

// MissingMessagesReq requests missing translation diagnostics for one locale.
type MissingMessagesReq struct {
	g.Meta    `path:"/i18n/messages/missing" method:"get" tags:"国际化" summary:"检查缺失翻译" dc:"检查目标语言相对当前默认语言缺失的翻译键，返回默认语言基线值和来源范围，供交付项目补齐翻译资源" permission:"system:i18n:diagnose"`
	Locale    string `json:"locale" dc:"目标语言编码，不传时按请求上下文自动解析，如 zh-CN、en-US" eg:"en-US"`
	KeyPrefix string `json:"keyPrefix" dc:"按翻译键前缀过滤，如 menu.、plugin.demo.；不传则检查全部" eg:"menu."`
}

// MissingMessageItem describes one missing translation key.
type MissingMessageItem struct {
	Key          string `json:"key" dc:"缺失的翻译键" eg:"menu.dashboard.title"`
	DefaultValue string `json:"defaultValue" dc:"默认语言中的基线翻译值" eg:"工作台"`
	SourceType   string `json:"sourceType" dc:"基线来源类型：host_file=宿主文件 plugin_file=插件文件 database=数据库覆写" eg:"host_file"`
	SourceKey    string `json:"sourceKey" dc:"基线来源标识，如 core 或具体 plugin_id" eg:"core"`
}

// MissingMessagesRes returns missing translation diagnostics.
type MissingMessagesRes struct {
	Locale        string               `json:"locale" dc:"目标语言编码" eg:"en-US"`
	DefaultLocale string               `json:"defaultLocale" dc:"当前宿主默认语言编码" eg:"zh-CN"`
	Total         int                  `json:"total" dc:"缺失翻译键数量" eg:"12"`
	Items         []MissingMessageItem `json:"items" dc:"缺失翻译键列表" eg:"[]"`
}
