package v1

// LocaleDirection identifies the text direction for one runtime locale.
type LocaleDirection string

const (
	// LocaleDirectionLTR means left-to-right text layout.
	LocaleDirectionLTR LocaleDirection = "ltr"
)

// MessageSourceType identifies the resource layer that supplied one message.
type MessageSourceType string

const (
	// MessageSourceTypeHostFile means the value came from host manifest files.
	MessageSourceTypeHostFile MessageSourceType = "host_file"
	// MessageSourceTypePluginFile means the value came from plugin manifest files.
	MessageSourceTypePluginFile MessageSourceType = "plugin_file"
)

// ExportMode identifies how a message export result was assembled.
type ExportMode string

const (
	// ExportModeEffective returns effective results with default-language fallback.
	ExportModeEffective ExportMode = "effective"
	// ExportModeRaw returns current-language raw resource values.
	ExportModeRaw ExportMode = "raw"
)
