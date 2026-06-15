// Package notifycap defines notification-domain capability contracts for
// plugins without exposing host notification tables.
package notifycap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
)

// MessageID identifies one notification message.
type MessageID string

// SourceType identifies the originating business source type for messages.
type SourceType string

// CategoryCode identifies the inbox category for messages.
type CategoryCode string

const (
	// SourceTypeNotice identifies notice-originated messages.
	SourceTypeNotice SourceType = "notice"
	// SourceTypePlugin identifies plugin-originated messages.
	SourceTypePlugin SourceType = "plugin"
)

const (
	// CategoryCodeOther identifies messages whose sender did not declare a category code.
	CategoryCodeOther CategoryCode = "other"
)

// SendInput defines one governed notification request.
type SendInput struct {
	// ChannelKey is the governed notification channel key authorized for this send.
	// Empty means adapters may use the built-in inbox channel when appropriate.
	ChannelKey string
	// Recipients contains target user domain IDs.
	Recipients []string
	// SourceType is the originating business source type.
	SourceType SourceType
	// SourceID is the originating business record identifier.
	SourceID string
	// Title is the message title.
	Title string
	// Content is the message content.
	Content string
	// Category is the plugin or host notification category.
	Category CategoryCode
	// Payload carries optional metadata stored with the notification.
	Payload map[string]any
	// SenderUserID is the optional sender user identifier. When zero, adapters
	// may use the actor in CapabilityContext.
	SenderUserID int64
}

// SendResult describes the created notification message.
type SendResult struct {
	// MessageID is the created message identifier.
	MessageID MessageID
	// DeliveryCount is the number of target deliveries.
	DeliveryCount int
}

// Service defines notification capability methods available to plugins.
type Service interface {
	// BatchGet returns visible message projections and opaque missing IDs.
	BatchGet(ctx context.Context, capCtx capmodel.CapabilityContext, ids []MessageID) (*capmodel.BatchResult[map[string]any, MessageID], error)
	// Send sends one governed notification message.
	Send(ctx context.Context, capCtx capmodel.CapabilityContext, input SendInput) (*SendResult, error)
}

// AdminService defines governed notification commands.
type AdminService interface {
	Service
	// Delete removes visible notification messages.
	Delete(ctx context.Context, capCtx capmodel.CapabilityContext, ids []MessageID) error
	// DeleteBySource removes visible notifications for business source IDs.
	DeleteBySource(ctx context.Context, capCtx capmodel.CapabilityContext, sourceType SourceType, sourceIDs []string) error
}

// ScopeService defines host-internal notification visibility helpers.
type ScopeService interface {
	// EnsureVisible rejects when any message is outside caller scope.
	EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []MessageID) error
}
