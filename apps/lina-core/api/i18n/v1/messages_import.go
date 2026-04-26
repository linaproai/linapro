// This file defines DTOs for importing flat i18n message overrides.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ImportMessagesReq requests importing flat translation messages into database overrides.
type ImportMessagesReq struct {
	g.Meta    `path:"/i18n/messages/import" method:"post" tags:"internationalization" summary:"Import internationalized messages" dc:"Batch import flat translation keys into the database override table by language, supporting different scopes such as host, project or plugin, suitable for batch revision of copywriting after the delivery project goes online" permission:"system:i18n:import"`
	Locale    string            `json:"locale" dc:"Target language encoding, such as zh-CN, en-US" eg:"en-US"`
	ScopeType string            `json:"scopeType" dc:"Scope type: host=host project=project plugin=plugin business=business" eg:"host"`
	ScopeKey  string            `json:"scopeKey" dc:"Scope identifier, such as core, project-code, plugin-id" eg:"core"`
	Overwrite bool              `json:"overwrite" dc:"Whether to overwrite when encountering an existing translation key under the same scope: true=overwrite false=skip" eg:"true"`
	Remark    string            `json:"remark" dc:"Notes are imported this time to facilitate subsequent diagnosis and tracking." eg:"bulk import from delivery pack"`
	Messages  map[string]string `json:"messages" dc:"A collection of translation key values organized by flat key" eg:"{}"`
}

// ImportMessagesRes returns the import summary.
type ImportMessagesRes struct {
	Locale   string `json:"locale" dc:"The target language encoding of this import" eg:"en-US"`
	Imported int    `json:"imported" dc:"The total number of translation keys processed in the request" eg:"64"`
	Created  int    `json:"created" dc:"Number of new translation keys" eg:"40"`
	Updated  int    `json:"updated" dc:"Override the number of updated translation keys" eg:"20"`
	Skipped  int    `json:"skipped" dc:"Number of translation keys skipped due to turning off overrides" eg:"4"`
}
