// This file defines DTOs for importing flat i18n message overrides.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ImportMessagesReq requests importing flat translation messages into database overrides.
type ImportMessagesReq struct {
	g.Meta    `path:"/i18n/messages/import" method:"post" tags:"国际化" summary:"导入国际化消息" dc:"按语言将扁平翻译键批量导入到数据库覆写表，支持宿主、项目或插件等不同作用域，适用于交付项目上线后批量修订文案" permission:"system:i18n:import"`
	Locale    string            `json:"locale" dc:"目标语言编码，如 zh-CN、en-US" eg:"en-US"`
	ScopeType string            `json:"scopeType" dc:"作用域类型：host=宿主 project=项目 plugin=插件 business=业务" eg:"host"`
	ScopeKey  string            `json:"scopeKey" dc:"作用域标识，如 core、project-code、plugin-id" eg:"core"`
	Overwrite bool              `json:"overwrite" dc:"遇到同 scope 下已存在的翻译键时是否覆盖：true=覆盖 false=跳过" eg:"true"`
	Remark    string            `json:"remark" dc:"本次导入备注，便于后续诊断与追踪" eg:"bulk import from delivery pack"`
	Messages  map[string]string `json:"messages" dc:"按扁平 key 组织的翻译键值集合" eg:"{}"`
}

// ImportMessagesRes returns the import summary.
type ImportMessagesRes struct {
	Locale   string `json:"locale" dc:"本次导入的目标语言编码" eg:"en-US"`
	Imported int    `json:"imported" dc:"请求中处理的翻译键总数" eg:"64"`
	Created  int    `json:"created" dc:"新建翻译键数量" eg:"40"`
	Updated  int    `json:"updated" dc:"覆盖更新的翻译键数量" eg:"20"`
	Skipped  int    `json:"skipped" dc:"因关闭覆盖而跳过的翻译键数量" eg:"4"`
}
