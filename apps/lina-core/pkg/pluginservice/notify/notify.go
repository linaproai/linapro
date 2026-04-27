// Package notify exposes a narrowed host notify contract to source plugins so
// they can publish notice messages and clean up delivered records without
// depending on host-internal service packages.
package notify

import (
	"context"

	internalnotify "lina-core/internal/service/notify"
)

// SourceType aliases the host notify source type enumeration.
type SourceType = internalnotify.SourceType

// CategoryCode aliases the host notify inbox category code enumeration.
type CategoryCode = internalnotify.CategoryCode

// NoticePublishInput aliases the host notice publication input contract.
type NoticePublishInput = internalnotify.NoticePublishInput

// SendOutput aliases the host notify send result contract.
type SendOutput = internalnotify.SendOutput

// Published notify source-type constants.
const (
	// SourceTypeNotice identifies notice-originated messages.
	SourceTypeNotice = internalnotify.SourceTypeNotice
	// SourceTypePlugin identifies plugin-originated messages.
	SourceTypePlugin = internalnotify.SourceTypePlugin
)

// Published notify inbox category-code constants. Plugins declare their own
// category codes as opaque strings; the host only publishes the generic
// fallback so that callers can default to it when no code is specified.
const (
	// CategoryCodeOther identifies inbox messages whose sender did not declare a category code.
	CategoryCodeOther = internalnotify.CategoryCodeOther
)

// Service defines the notify operations published to source plugins.
type Service interface {
	// SendNoticePublication fans one published notice into the host inbox pipeline.
	SendNoticePublication(ctx context.Context, in NoticePublishInput) (*SendOutput, error)
	// DeleteBySource removes host notify deliveries and messages for the given business source identifiers.
	DeleteBySource(ctx context.Context, sourceType SourceType, sourceIDs []string) error
}

// serviceAdapter bridges the internal notify service into the published plugin contract.
type serviceAdapter struct {
	service internalnotify.Service
}

// New creates and returns the published notify service adapter.
func New() Service {
	return &serviceAdapter{service: internalnotify.New()}
}

// SendNoticePublication fans one published notice into the host inbox pipeline.
func (s *serviceAdapter) SendNoticePublication(ctx context.Context, in NoticePublishInput) (*SendOutput, error) {
	if s == nil || s.service == nil {
		return nil, nil
	}
	return s.service.SendNoticePublication(ctx, in)
}

// DeleteBySource removes host notify deliveries and messages for the given business source identifiers.
func (s *serviceAdapter) DeleteBySource(ctx context.Context, sourceType SourceType, sourceIDs []string) error {
	if s == nil || s.service == nil {
		return nil
	}
	return s.service.DeleteBySource(ctx, sourceType, sourceIDs)
}
