// This file defines DTOs for the runtime i18n locale-list API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// RuntimeLocalesReq requests the runtime locale list exposed by the host.
type RuntimeLocalesReq struct {
	g.Meta `path:"/i18n/runtime/locales" method:"get" tags:"国际化" summary:"获取运行时语言列表" dc:"返回宿主当前支持的运行时语言列表，包含语言编码、默认标记、展示名称和原生名称，供登录页、工作台或交付工具构建语言切换器"`
	Lang   string `json:"lang" dc:"名称展示使用的语言编码；不传时按请求上下文自动解析，如 zh-CN、en-US" eg:"en-US"`
}

// RuntimeLocaleItem describes one runtime locale option.
type RuntimeLocaleItem struct {
	Locale     string `json:"locale" dc:"语言编码，如 zh-CN、en-US" eg:"zh-CN"`
	Name       string `json:"name" dc:"按当前展示语言本地化后的语言名称" eg:"Chinese (Simplified)"`
	NativeName string `json:"nativeName" dc:"该语言自己的原生名称" eg:"简体中文"`
	IsDefault  bool   `json:"isDefault" dc:"是否为宿主默认语言：true=是 false=否" eg:"true"`
}

// RuntimeLocalesRes returns the runtime locale list payload.
type RuntimeLocalesRes struct {
	Locale string              `json:"locale" dc:"本次请求最终生效的展示语言编码" eg:"en-US"`
	Items  []RuntimeLocaleItem `json:"items" dc:"宿主当前支持的运行时语言列表" eg:"[]"`
}
