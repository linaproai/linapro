// This file defines typed runtime i18n values shared by locale metadata and
// resource source diagnostics.

package i18n

// LocaleDirection represents the document text direction for one locale.
type LocaleDirection string

const (
	// LocaleDirectionLTR means left-to-right text layout.
	LocaleDirectionLTR LocaleDirection = "ltr"
)

// String returns the persisted string value for the locale direction.
func (direction LocaleDirection) String() string {
	return string(direction)
}

// messageScopeType represents the logical owner scope of one i18n resource.
type messageScopeType string

const (
	// messageScopeTypeHost means the message belongs to the host scope.
	messageScopeTypeHost messageScopeType = "host"
	// messageScopeTypePlugin means the message belongs to one plugin scope.
	messageScopeTypePlugin messageScopeType = "plugin"
)

// messageOriginType represents the effective source layer of one runtime message.
type messageOriginType string

const (
	// messageOriginTypeHostFile means the message came from host manifest files.
	messageOriginTypeHostFile messageOriginType = "host_file"
	// messageOriginTypePluginFile means the message came from plugin manifest files.
	messageOriginTypePluginFile messageOriginType = "plugin_file"
)

const (
	// hostMessageScopeKey is the stable host scope key used for host-owned resources.
	hostMessageScopeKey = "core"
)
