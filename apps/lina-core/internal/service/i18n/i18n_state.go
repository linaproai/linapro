// This file defines typed runtime i18n status values used across locale and
// database-backed message loading.

package i18n

// localeStatus represents the enablement status stored in sys_i18n_locale.
type localeStatus int

const (
	// localeStatusDisabled means the locale is not exposed by the runtime host.
	localeStatusDisabled localeStatus = 0
	// localeStatusEnabled means the locale is enabled for runtime use.
	localeStatusEnabled localeStatus = 1
)

// localeDefaultFlag represents the default-language flag stored in sys_i18n_locale.
type localeDefaultFlag int

const (
	// localeDefaultNo means the locale is not the runtime default.
	localeDefaultNo localeDefaultFlag = 0
	// localeDefaultYes means the locale is the runtime default.
	localeDefaultYes localeDefaultFlag = 1
)

// messageStatus represents the enablement status stored in sys_i18n_message.
type messageStatus int

const (
	// messageStatusDisabled means the database override is disabled.
	messageStatusDisabled messageStatus = 0
	// messageStatusEnabled means the database override is enabled.
	messageStatusEnabled messageStatus = 1
)

// messageScopeType represents the logical scope recorded for one i18n message row.
type messageScopeType string

const (
	// messageScopeTypeHost means the message belongs to the host scope.
	messageScopeTypeHost messageScopeType = "host"
	// messageScopeTypeProject means the message belongs to one project delivery scope.
	messageScopeTypeProject messageScopeType = "project"
	// messageScopeTypePlugin means the message belongs to one plugin scope.
	messageScopeTypePlugin messageScopeType = "plugin"
	// messageScopeTypeBusiness means the message belongs to one business content scope.
	messageScopeTypeBusiness messageScopeType = "business"
)

// messageSourceType represents how one database message override was produced.
type messageSourceType string

const (
	// messageSourceTypeManual means the message was maintained manually.
	messageSourceTypeManual messageSourceType = "manual"
	// messageSourceTypeImport means the message was written by an import operation.
	messageSourceTypeImport messageSourceType = "import"
	// messageSourceTypeSync means the message was synchronized from another source.
	messageSourceTypeSync messageSourceType = "sync"
)

// messageOriginType represents the effective source layer of one runtime message.
type messageOriginType string

const (
	// messageOriginTypeHostFile means the message came from host manifest files.
	messageOriginTypeHostFile messageOriginType = "host_file"
	// messageOriginTypePluginFile means the message came from plugin manifest files.
	messageOriginTypePluginFile messageOriginType = "plugin_file"
	// messageOriginTypeDatabase means the message came from sys_i18n_message.
	messageOriginTypeDatabase messageOriginType = "database"
)

const (
	// hostMessageScopeKey is the stable host scope key used for host-owned overrides.
	hostMessageScopeKey = "core"
)
