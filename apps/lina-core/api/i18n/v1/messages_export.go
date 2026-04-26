// This file defines DTOs for exporting flat i18n runtime messages.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ExportMessagesReq requests one flat runtime message export for the given locale.
type ExportMessagesReq struct {
	g.Meta `path:"/i18n/messages/export" method:"get" tags:"internationalization" summary:"Export internationalized messages" dc:"Export a flat internationalized message collection in a specified language, supporting the export of effective aggregation results or original resources in the current language for delivery maintenance, offline proofreading and re-importing" permission:"system:i18n:export"`
	Locale string `json:"locale" dc:"Target language encoding, automatically parsed according to request context if not passed, such as zh-CN, en-US" eg:"en-US"`
	Raw    bool   `json:"raw" dc:"Whether to export only original resources in the current language: true=Export only aggregated results in the current language false=Export valid results with default language fallback" eg:"false"`
}

// ExportMessagesRes returns one flat runtime message export payload.
type ExportMessagesRes struct {
	Locale        string            `json:"locale" dc:"The target language encoding for this export" eg:"en-US"`
	DefaultLocale string            `json:"defaultLocale" dc:"Current host default language encoding" eg:"zh-CN"`
	Mode          string            `json:"mode" dc:"Export mode: effective=effective results with default language fallback raw=current language raw resources" eg:"effective"`
	Total         int               `json:"total" dc:"Number of translation keys exported" eg:"128"`
	Messages      map[string]string `json:"messages" dc:"A collection of internationalized messages output by flat key" eg:"{}"`
}
