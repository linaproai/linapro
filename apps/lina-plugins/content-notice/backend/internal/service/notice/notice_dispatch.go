// notice_dispatch.go implements the publication fan-out helpers that forward
// newly published notices into the host inbox pipeline through the notify
// bridge exposed to source plugins.

package notice

import (
	"context"

	"lina-core/pkg/pluginservice/notify"
)

// dispatchPublishedNotice delivers one published notice into the unified inbox
// pipeline after the notice record is persisted.
func (s *serviceImpl) dispatchPublishedNotice(
	ctx context.Context,
	noticeID int64,
	title string,
	content string,
	noticeType int,
	senderUserID int64,
) error {
	_, err := s.notifySvc.SendNoticePublication(ctx, notify.NoticePublishInput{
		NoticeID:     noticeID,
		Title:        title,
		Content:      content,
		CategoryCode: s.noticeTypeToCategoryCode(noticeType),
		SenderUserID: senderUserID,
	})
	return err
}

// noticeTypeToCategoryCode maps notice types to notify inbox category codes.
func (s *serviceImpl) noticeTypeToCategoryCode(noticeType int) notify.CategoryCode {
	switch noticeType {
	case NoticeTypeAnnouncement:
		return notify.CategoryCodeAnnouncement
	default:
		return notify.CategoryCodeNotice
	}
}
