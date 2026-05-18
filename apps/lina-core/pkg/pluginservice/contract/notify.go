// This file defines the source-plugin visible notification contract.

package contract

import "context"

// SourceType identifies the originating business source type for messages.
type SourceType string

// CategoryCode identifies the inbox category for messages.
type CategoryCode string

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

// SendOutput defines one unified notification send result.
type SendOutput struct {
	// MessageID is the created notify message identifier.
	MessageID int64
	// DeliveryCount is the number of created delivery rows.
	DeliveryCount int
}

// Published notify source-type constants.
const (
	// SourceTypeNotice identifies notice-originated messages.
	SourceTypeNotice SourceType = "notice"
	// SourceTypePlugin identifies plugin-originated messages.
	SourceTypePlugin SourceType = "plugin"
)

// Published notify inbox category-code constants. Plugins declare their own
// category codes as opaque strings; the host only publishes the generic
// fallback so that callers can default to it when no code is specified.
const (
	// CategoryCodeOther identifies inbox messages whose sender did not declare a category code.
	CategoryCodeOther CategoryCode = "other"
)

// NotifyService defines the notify operations published to source plugins.
type NotifyService interface {
	// SendNoticePublication fans one published notice into the host inbox pipeline.
	SendNoticePublication(ctx context.Context, in NoticePublishInput) (*SendOutput, error)
	// DeleteBySource removes host notify records for the given business source identifiers.
	DeleteBySource(ctx context.Context, sourceType SourceType, sourceIDs []string) error
}
