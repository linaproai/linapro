// This file defines DTOs for exporting flat i18n runtime messages.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ExportMessagesReq requests one flat runtime message export for the given locale.
type ExportMessagesReq struct {
	g.Meta `path:"/i18n/messages/export" method:"get" tags:"internationalization" summary:"Export internationalized messages" dc:"Export a flat internationalized message collection in a specified language for delivery maintenance and offline proofreading, with changes expected to be written back to JSON resource files" permission:"system:i18n:export"`
	Locale string `json:"locale" dc:"Target language encoding, automatically parsed according to request context if not passed, such as zh-CN, en-US" eg:"en-US"`
	Raw    bool   `json:"raw" dc:"Reserved export mode flag, currently both true and false return the resource-backed merged catalog" eg:"false"`
}

// ExportMessagesRes returns one flat runtime message export payload.
type ExportMessagesRes struct {
	Locale        string            `json:"locale" dc:"The target language encoding for this export" eg:"en-US"`
	DefaultLocale string            `json:"defaultLocale" dc:"Current host default language encoding" eg:"zh-CN"`
	Mode          string            `json:"mode" dc:"Export mode: effective=effective results with default language fallback raw=current language raw resources" eg:"effective"`
	Total         int               `json:"total" dc:"Number of translation keys exported" eg:"128"`
	Messages      map[string]string `json:"messages" dc:"A collection of internationalized messages output by flat key" eg:"{}"`
}
