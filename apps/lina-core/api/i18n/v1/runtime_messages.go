// This file defines DTOs for the runtime i18n message-bundle API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// RuntimeMessagesReq requests the aggregated runtime translation bundle for one locale.
type RuntimeMessagesReq struct {
	g.Meta `path:"/i18n/runtime/messages" method:"get" tags:"internationalization" summary:"Get the runtime internationalization message package" dc:"Returns the runtime internationalized message package aggregated between the host and the accessed resources for loading by the login page, management workbench, and host embedded plugin page."`
	Lang   string `json:"lang" dc:"Target language encoding; if not passed, it will be automatically parsed according to the request context, such as zh-CN, en-US" eg:"en-US"`
}

// RuntimeMessagesRes returns one runtime translation bundle.
type RuntimeMessagesRes struct {
	Locale   string                 `json:"locale" dc:"The final effective language encoding of this request" eg:"en-US"`
	Messages map[string]interface{} `json:"messages" dc:"Aggregated runtime internationalized message collection" eg:"{}"`
}
