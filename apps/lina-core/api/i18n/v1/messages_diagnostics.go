// This file defines DTOs for runtime i18n source diagnostics.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DiagnoseMessagesReq requests source diagnostics for one locale.
type DiagnoseMessagesReq struct {
	g.Meta    `path:"/i18n/messages/diagnostics" method:"get" tags:"internationalization" summary:"diagnostic translation source" dc:"Diagnose the actual hit source of each translation key in the target language, whether it falls back to the default language, and the final effective language to facilitate troubleshooting overwrite priority and missing translation issues." permission:"system:i18n:diagnose"`
	Locale    string `json:"locale" dc:"Target language encoding, automatically parsed according to request context if not passed, such as zh-CN, en-US" eg:"en-US"`
	KeyPrefix string `json:"keyPrefix" dc:"Filter by translation key prefix, such as menu., plugin.demo.; if not passed, all will be diagnosed." eg:"menu."`
}

// MessageDiagnosticItem describes the effective resolution result for one translation key.
type MessageDiagnosticItem struct {
	Key             string `json:"key" dc:"Translation key" eg:"menu.dashboard.title"`
	Value           string `json:"value" dc:"The current final effective translation value" eg:"Workbench"`
	RequestedLocale string `json:"requestedLocale" dc:"Requested target language encoding" eg:"en-US"`
	EffectiveLocale string `json:"effectiveLocale" dc:"The language encoding that actually provides this translation value" eg:"en-US"`
	FromFallback    bool   `json:"fromFallback" dc:"Whether to fall back to the default language: true=yes false=no" eg:"false"`
	SourceType      string `json:"sourceType" dc:"Hit source type: host_file=host manifest file plugin_file=plugin manifest file or dynamic plugin asset" eg:"host_file"`
	SourceKey       string `json:"sourceKey" dc:"Hit source identifier, such as core, plugin-id, or specific scope key" eg:"core"`
}

// DiagnoseMessagesRes returns runtime source diagnostics for one locale.
type DiagnoseMessagesRes struct {
	Locale        string                  `json:"locale" dc:"target language encoding" eg:"en-US"`
	DefaultLocale string                  `json:"defaultLocale" dc:"Current host default language encoding" eg:"zh-CN"`
	Total         int                     `json:"total" dc:"Number of translation keys returned by diagnostics" eg:"128"`
	Items         []MessageDiagnosticItem `json:"items" dc:"Translation source diagnostic result list" eg:"[]"`
}
