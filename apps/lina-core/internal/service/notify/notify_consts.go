// This file defines shared notify domain enums and conversion helpers.

package notify

// ChannelType defines the supported notify channel types.
type ChannelType string

// SourceType defines the supported notify source types.
type SourceType string

// CategoryCode defines the supported notify category codes.
type CategoryCode string

// RecipientType defines the supported notify recipient types.
type RecipientType string

const (
	// ChannelKeyInbox is the built-in inbox channel key.
	ChannelKeyInbox = "inbox"
)

const (
	// ChannelTypeInbox identifies inbox deliveries.
	ChannelTypeInbox ChannelType = "inbox"
	// ChannelTypeEmail identifies email deliveries.
	ChannelTypeEmail ChannelType = "email"
	// ChannelTypeWebhook identifies webhook deliveries.
	ChannelTypeWebhook ChannelType = "webhook"
)

const (
	// SourceTypeNotice identifies notice-originated messages.
	SourceTypeNotice SourceType = "notice"
	// SourceTypePlugin identifies plugin-originated messages.
	SourceTypePlugin SourceType = "plugin"
	// SourceTypeSystem identifies system-originated messages.
	SourceTypeSystem SourceType = "system"
)

const (
	// CategoryCodeNotice identifies notice messages.
	CategoryCodeNotice CategoryCode = "notice"
	// CategoryCodeAnnouncement identifies announcement messages.
	CategoryCodeAnnouncement CategoryCode = "announcement"
	// CategoryCodeOther identifies all other inbox messages.
	CategoryCodeOther CategoryCode = "other"
)

const (
	// RecipientTypeUser identifies inbox user recipients.
	RecipientTypeUser RecipientType = "user"
	// RecipientTypeEmail identifies email recipients.
	RecipientTypeEmail RecipientType = "email"
	// RecipientTypeWebhook identifies webhook recipients.
	RecipientTypeWebhook RecipientType = "webhook"
)

const (
	// ChannelStatusDisabled marks a disabled channel row.
	ChannelStatusDisabled = 0
	// ChannelStatusEnabled marks an enabled channel row.
	ChannelStatusEnabled = 1
)

const (
	// DeliveryStatusPending marks a queued delivery.
	DeliveryStatusPending = 0
	// DeliveryStatusSucceeded marks a successful delivery.
	DeliveryStatusSucceeded = 1
	// DeliveryStatusFailed marks a failed delivery.
	DeliveryStatusFailed = 2
)

const (
	// LegacyMessageTypeNotice maps inbox rows to the existing user message notice type.
	LegacyMessageTypeNotice = 1
	// LegacyMessageTypeAnnouncement maps inbox rows to the existing user message announcement type.
	LegacyMessageTypeAnnouncement = 2
)

// String returns the canonical channel type value.
func (value ChannelType) String() string {
	return string(value)
}

// String returns the canonical source type value.
func (value SourceType) String() string {
	return string(value)
}

// String returns the canonical category code value.
func (value CategoryCode) String() string {
	return string(value)
}

// String returns the canonical recipient type value.
func (value RecipientType) String() string {
	return string(value)
}

func categoryCodeToLegacyMessageType(categoryCode CategoryCode) int {
	switch categoryCode {
	case CategoryCodeAnnouncement:
		return LegacyMessageTypeAnnouncement
	default:
		return LegacyMessageTypeNotice
	}
}
