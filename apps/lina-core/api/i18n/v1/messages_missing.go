// This file defines DTOs for checking missing i18n messages.

package v1

import "github.com/gogf/gf/v2/frame/g"

// MissingMessagesReq requests missing translation diagnostics for one locale.
type MissingMessagesReq struct {
	g.Meta    `path:"/i18n/messages/missing" method:"get" tags:"internationalization" summary:"Check for missing translations" dc:"Check the missing translation keys in the target language relative to the current default language, and return the default language baseline value and source range for the delivery project to complete the translation resources." permission:"system:i18n:diagnose"`
	Locale    string `json:"locale" dc:"Target language encoding, automatically parsed according to request context if not passed, such as zh-CN, en-US" eg:"en-US"`
	KeyPrefix string `json:"keyPrefix" dc:"Filter by translation key prefix, such as menu., plugin.demo.; if not passed, check all" eg:"menu."`
}

// MissingMessageItem describes one missing translation key.
type MissingMessageItem struct {
	Key          string `json:"key" dc:"Missing translation key" eg:"menu.dashboard.title"`
	DefaultValue string `json:"defaultValue" dc:"Baseline translation value in default language" eg:"workbench"`
	SourceType   string `json:"sourceType" dc:"Baseline source type: host_file=host manifest file plugin_file=plugin manifest file or dynamic plugin asset" eg:"host_file"`
	SourceKey    string `json:"sourceKey" dc:"Baseline source ID, such as core or specific plugin_id" eg:"core"`
}

// MissingMessagesRes returns missing translation diagnostics.
type MissingMessagesRes struct {
	Locale        string               `json:"locale" dc:"target language encoding" eg:"en-US"`
	DefaultLocale string               `json:"defaultLocale" dc:"Current host default language encoding" eg:"zh-CN"`
	Total         int                  `json:"total" dc:"Number of missing translation keys" eg:"12"`
	Items         []MissingMessageItem `json:"items" dc:"Missing translation key list" eg:"[]"`
}
