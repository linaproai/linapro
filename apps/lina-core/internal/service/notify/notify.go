// Package notify provides unified notification send, inbox query, and notice
// publication services for host modules and plugins.
package notify

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"
)

// Service defines the notify service contract.
type Service interface {
	// InboxUnreadCount returns the unread inbox delivery count for one user.
	InboxUnreadCount(ctx context.Context, userID int64) (int, error)
	// InboxList returns one paged inbox list for the current user.
	InboxList(ctx context.Context, in InboxListInput) (*InboxListOutput, error)
	// InboxMarkRead marks one inbox delivery as read for the current user.
	InboxMarkRead(ctx context.Context, userID int64, deliveryID int64) error
	// InboxMarkAllRead marks all unread inbox deliveries as read for the current user.
	InboxMarkAllRead(ctx context.Context, userID int64) error
	// InboxDelete soft-deletes one inbox delivery for the current user.
	InboxDelete(ctx context.Context, userID int64, deliveryID int64) error
	// InboxClear soft-deletes all inbox deliveries for the current user.
	InboxClear(ctx context.Context, userID int64) error
	// DeleteBySource removes notify deliveries and messages for the given business source identifiers.
	DeleteBySource(ctx context.Context, sourceType SourceType, sourceIDs []string) error
	// Send validates the notify channel and creates unified notify message and delivery records.
	Send(ctx context.Context, in SendInput) (*SendOutput, error)
	// SendNoticePublication sends one published notice through the built-in inbox channel.
	SendNoticePublication(ctx context.Context, in NoticePublishInput) (*SendOutput, error)
}

// Interface compliance assertion for the default notify service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

// SendInput defines one unified notification send request.
type SendInput struct {
	// ChannelKey is the target notify channel identifier.
	ChannelKey string

	// PluginID is the optional originating plugin identifier.
	PluginID string

	// SourceType is the originating business source type.
	SourceType SourceType

	// SourceID is the originating business record identifier.
	SourceID string

	// CategoryCode is the inbox category mapped to the outgoing message.
	CategoryCode CategoryCode

	// Title is the visible message title.
	Title string

	// Content is the visible message body.
	Content string
	// Payload carries optional structured message metadata.
	Payload map[string]any
	// SenderUserID is the optional sender user identifier.
	SenderUserID int64
	// RecipientUserIDs is the ordered recipient user identifier list for inbox delivery.
	RecipientUserIDs []int64
}

// SendOutput defines one unified notification send result.
type SendOutput struct {
	// MessageID is the created notify message identifier.
	MessageID int64
	// DeliveryCount is the number of created delivery rows.
	DeliveryCount int
}

// NoticePublishInput defines one notice publication fan-out request.
type NoticePublishInput struct {
	// NoticeID is the published notice identifier.
	NoticeID int64
	// Title is the notice title.
	Title string
	// Content is the notice body content.
	Content string
	// CategoryCode is the inbox category mapped from notice type.
	CategoryCode CategoryCode
	// SenderUserID is the user who created or published the notice.
	SenderUserID int64
}

// InboxListInput defines the inbox list query input.
type InboxListInput struct {
	// UserID is the current inbox user identifier.
	UserID int64
	// PageNum is the 1-based page number.
	PageNum int
	// PageSize is the requested page size.
	PageSize int
}

// InboxListOutput defines the inbox list query result.
type InboxListOutput struct {
	// List is the ordered inbox message slice.
	List []*InboxListItem
	// Total is the total number of matching inbox rows before pagination.
	Total int
}

// InboxListItem defines one inbox list item exposed through the user message facade.
type InboxListItem struct {
	// Id is the notify delivery identifier exposed as the inbox message ID.
	Id int64
	// UserID is the inbox owner user identifier.
	UserID int64
	// Title is the message title displayed in the inbox.
	Title string
	// CategoryCode is the opaque sender-declared category code stored on the message.
	CategoryCode string
	// SourceType is the originating business source type.
	SourceType string
	// SourceID is the legacy numeric source identifier used by current previews.
	SourceID int64
	// IsRead reports whether the inbox row has been marked as read.
	IsRead int
	// ReadAt is the optional read timestamp.
	ReadAt *gtime.Time
	// CreatedAt is the inbox delivery creation timestamp.
	CreatedAt *gtime.Time
}

// New creates and returns a new notify service instance.
func New() Service {
	return &serviceImpl{}
}
