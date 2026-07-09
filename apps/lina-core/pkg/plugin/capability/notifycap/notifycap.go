// Package notifycap defines notification-domain capability contracts for
// plugins without exposing host notification tables.
package notifycap

import (
	"context"

	usermsgv1 "lina-core/api/usermsg/v1"
	"lina-core/pkg/plugin/capability/capmodel"
)

// Service defines governed notification capability methods available to
// plugins. Reads use bounded info queries and recipient visibility;
// sends and deletes must validate channel/resource authorization, tenant
// boundary, source ownership, audit source, and delivery side effects.
type Service interface {
	// Get returns one visible notification message info record.
	Get(ctx context.Context, id MessageID) (*MessageInfo, error)
	// BatchGet returns visible message info records and opaque missing IDs.
	BatchGet(ctx context.Context, ids []MessageID) (*capmodel.BatchResult[*MessageInfo, MessageID], error)
	// List returns one bounded page of visible notification messages.
	List(ctx context.Context, input ListInput) (*capmodel.PageResult[*MessageInfo], error)
	// BatchGetBySource returns visible message info records grouped by source ID.
	BatchGetBySource(ctx context.Context, input BatchGetBySourceInput) (*BatchGetBySourceResult, error)
	// EnsureVisible rejects when any requested message is absent or outside caller scope.
	EnsureVisible(ctx context.Context, ids []MessageID) error
	// Send sends one governed notification message.
	Send(ctx context.Context, input SendInput) (*SendResult, error)
	// Delete removes visible notification messages after recipient, tenant,
	// source, batch, audit, and cascade-boundary checks.
	Delete(ctx context.Context, ids []MessageID) error
	// DeleteBySource removes visible notifications for bounded business source
	// IDs after source ownership, tenant, audit, and cascade-boundary checks.
	DeleteBySource(ctx context.Context, sourceType usermsgv1.SourceType, sourceIDs []string) error
	// MarkRead marks visible notifications as read for the current actor.
	MarkRead(ctx context.Context, ids []MessageID) error
	// MarkUnread marks visible notifications as unread for the current actor.
	MarkUnread(ctx context.Context, ids []MessageID) error
}

// MessageID identifies one notification message.
type MessageID string

// CategoryCode identifies the inbox category for messages.
type CategoryCode string

const (
	// CategoryCodeOther identifies messages whose sender did not declare a category code.
	CategoryCodeOther CategoryCode = "other"
)

const (
	// MaxBatchGetMessages limits notification message batch reads.
	MaxBatchGetMessages = 100
	// MaxBatchGetBySourceIDs limits source IDs in one source batch read.
	MaxBatchGetBySourceIDs = 100
	// MaxBatchGetBySourceMessages limits visible messages returned by one source batch read.
	MaxBatchGetBySourceMessages = 200
	// MaxEnsureVisibleMessages limits message visibility checks.
	MaxEnsureVisibleMessages = 100
	// MaxListMessagesPageSize limits one notification message page.
	MaxListMessagesPageSize = 200
)

// MessageInfo describes one plugin-visible notification message without
// exposing notify storage internals or arbitrary extension fields.
type MessageInfo struct {
	// ID is the notification message identifier.
	ID MessageID `json:"id"`
	// TenantID is the owning tenant identifier; zero means platform scope.
	TenantID int `json:"tenantId"`
	// PluginID is the source plugin identifier when the message came from a plugin.
	PluginID string `json:"pluginId,omitempty"`
	// SourceType is the originating business source type.
	SourceType usermsgv1.SourceType `json:"sourceType"`
	// SourceID is the originating business source identifier.
	SourceID string `json:"sourceId"`
	// CategoryCode is the inbox category code.
	CategoryCode CategoryCode `json:"categoryCode"`
	// Title is the notification title.
	Title string `json:"title"`
	// CreatedAt is the creation timestamp as Unix milliseconds.
	CreatedAt int64 `json:"createdAt,omitempty"`
}

// BatchGetBySourceInput describes one source-based notification batch read.
type BatchGetBySourceInput struct {
	// SourceType is the originating business source type.
	SourceType usermsgv1.SourceType `json:"sourceType"`
	// SourceIDs contains source identifiers to read in a single bounded query.
	SourceIDs []string `json:"sourceIds"`
}

// BatchGetBySourceResult groups visible message info records by source ID.
type BatchGetBySourceResult struct {
	// Items stores visible messages keyed by source ID.
	Items map[string][]*MessageInfo `json:"items"`
	// MissingIDs stores source IDs with no visible messages.
	MissingIDs []string `json:"missingIds"`
}

// ListInput constrains notification message listing.
type ListInput struct {
	// SourceType optionally filters by originating business source.
	SourceType usermsgv1.SourceType `json:"sourceType,omitempty"`
	// SourceID optionally filters by originating business record.
	SourceID string `json:"sourceId,omitempty"`
	// CategoryCode optionally filters by inbox category.
	CategoryCode CategoryCode `json:"categoryCode,omitempty"`
	// Page constrains result size and stable sorting.
	Page capmodel.PageRequest `json:"page"`
}

// SendInput defines one governed notification request.
type SendInput struct {
	// ChannelKey is the governed notification channel key authorized for this send.
	// Empty means adapters may use the built-in inbox channel when appropriate.
	ChannelKey string
	// Recipients contains target user domain IDs.
	Recipients []string
	// SourceType is the originating business source type.
	SourceType usermsgv1.SourceType
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
	// may use the current user from the standard business context.
	SenderUserID int64
}

// SendResult describes the created notification message.
type SendResult struct {
	// MessageID is the created message identifier.
	MessageID MessageID
	// DeliveryCount is the number of target deliveries.
	DeliveryCount int
}
