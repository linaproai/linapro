// This file defines DTOs for the runtime i18n locale-list API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// RuntimeLocalesReq requests the runtime locale list exposed by the host.
type RuntimeLocalesReq struct {
	g.Meta `path:"/i18n/runtime/locales" method:"get" tags:"internationalization" summary:"Get the runtime language list" dc:"Returns a list of runtime languages currently supported by the host, including language encoding, default markup, display name and native name, for use by landing pages, workbench or delivery tools to build language switchers"`
	Lang   string `json:"lang" dc:"The language encoding used for name display; if not passed, it will be automatically parsed according to the request context, such as zh-CN, en-US" eg:"en-US"`
}

// RuntimeLocaleItem describes one runtime locale option.
type RuntimeLocaleItem struct {
	Locale     string `json:"locale" dc:"Language encoding, such as zh-CN, en-US" eg:"zh-CN"`
	Name       string `json:"name" dc:"Language name localized according to the current display language" eg:"Chinese (Simplified)"`
	NativeName string `json:"nativeName" dc:"The language's own native name" eg:"Simplified Chinese"`
	IsDefault  bool   `json:"isDefault" dc:"Whether it is the default language of the host: true=yes false=no" eg:"true"`
}

// RuntimeLocalesRes returns the runtime locale list payload.
type RuntimeLocalesRes struct {
	Locale string              `json:"locale" dc:"The final effective display language encoding of this request" eg:"en-US"`
	Items  []RuntimeLocaleItem `json:"items" dc:"List of runtime languages currently supported by the host" eg:"[]"`
}
