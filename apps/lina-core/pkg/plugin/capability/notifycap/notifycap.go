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

// Service defines read-oriented notification capability methods.
type Service interface {
	// BatchGetMessages returns visible message projections and opaque missing IDs.
	BatchGetMessages(ctx context.Context, capCtx capmodel.CapabilityContext, ids []MessageID) (*capmodel.BatchResult[map[string]any, MessageID], error)
}

// AdminService defines governed notification commands.
type AdminService interface {
	// Send sends one governed notification message.
	Send(ctx context.Context, capCtx capmodel.CapabilityContext, input SendInput) (*SendResult, error)
	// DeleteMessages removes visible notification messages.
	DeleteMessages(ctx context.Context, capCtx capmodel.CapabilityContext, ids []MessageID) error
	// DeleteBySource removes visible notifications for business source IDs.
	DeleteBySource(ctx context.Context, capCtx capmodel.CapabilityContext, sourceType SourceType, sourceIDs []string) error
}

// ScopeService defines host-internal notification visibility helpers.
type ScopeService interface {
	// EnsureMessagesVisible rejects when any message is outside caller scope.
	EnsureMessagesVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []MessageID) error
}
