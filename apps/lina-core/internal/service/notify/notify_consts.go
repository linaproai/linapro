// This file defines shared notify domain enums and conversion helpers.

package notify

import usermsgv1 "lina-core/api/usermsg/v1"

// ChannelType defines the supported notify channel types.
type ChannelType string

// SourceType reuses the user-message API source enum for notify origins.
type SourceType = usermsgv1.SourceType

// CategoryCode defines the supported notify category codes.
type CategoryCode string

// RecipientType defines the supported notify recipient types.
type RecipientType string

// Channel key constants identify built-in notify transport channels.
const (
	// ChannelKeyInbox is the built-in inbox channel key.
	ChannelKeyInbox = "inbox"
)

// Channel type constants enumerate the supported notify transport families.
const (
	// ChannelTypeInbox identifies inbox deliveries.
	ChannelTypeInbox ChannelType = "inbox"
	// ChannelTypeEmail identifies email deliveries.
	ChannelTypeEmail ChannelType = "email"
	// ChannelTypeWebhook identifies webhook deliveries.
	ChannelTypeWebhook ChannelType = "webhook"
)

// Category code constants enumerate the inbox message categories owned by the
// host itself. Senders (such as source plugins) are expected to declare their
// own category codes as opaque strings; the host does not enumerate plugin
// categories here. CategoryCodeOther is the canonical fallback used when an
// inbound send omits the category code.
const (
	// CategoryCodeSystem identifies notifications produced by the host system itself.
	CategoryCodeSystem CategoryCode = "system"
	// CategoryCodeOther identifies all other inbox messages whose sender did not declare a category code.
	CategoryCodeOther CategoryCode = "other"
)

// Recipient type constants enumerate the supported delivery recipient kinds.
const (
	// RecipientTypeUser identifies inbox user recipients.
	RecipientTypeUser RecipientType = "user"
	// RecipientTypeEmail identifies email recipients.
	RecipientTypeEmail RecipientType = "email"
	// RecipientTypeWebhook identifies webhook recipients.
	RecipientTypeWebhook RecipientType = "webhook"
)

// Delivery status constants reflect persisted notify delivery outcomes.
const (
	// DeliveryStatusPending marks a queued delivery.
	DeliveryStatusPending = 0
	// DeliveryStatusSucceeded marks a successful delivery.
	DeliveryStatusSucceeded = 1
	// DeliveryStatusFailed marks a failed delivery.
	DeliveryStatusFailed = 2
)

// String returns the canonical channel type value.
func (value ChannelType) String() string {
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
