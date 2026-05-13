// notice_dispatch.go implements the publication fan-out helpers that forward
// newly published notices into the host inbox pipeline through the notify
// bridge exposed to source plugins.

package notice

import (
	"context"

	plugincontract "lina-core/pkg/pluginservice/contract"
)

// Plugin-local notify category codes. The host notify service treats category
// codes as opaque sender-declared strings; this plugin owns its own category
// vocabulary and registers matching translations under
// `notify.category.{code}.label` / `.color` in its own manifest/i18n bundles.
const (
	// noticeCategoryCodeNotice is the opaque category code for general notice messages dispatched by this plugin.
	noticeCategoryCodeNotice plugincontract.CategoryCode = "notice"
	// noticeCategoryCodeAnnouncement is the opaque category code for announcement messages dispatched by this plugin.
	noticeCategoryCodeAnnouncement plugincontract.CategoryCode = "announcement"
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
	_, err := s.notifySvc.SendNoticePublication(ctx, plugincontract.NoticePublishInput{
		NoticeID:     noticeID,
		Title:        title,
		Content:      content,
		CategoryCode: s.noticeTypeToCategoryCode(noticeType),
		SenderUserID: senderUserID,
	})
	return err
}

// noticeTypeToCategoryCode maps plugin-owned notice types to plugin-owned
// notify inbox category codes.
func (s *serviceImpl) noticeTypeToCategoryCode(noticeType int) plugincontract.CategoryCode {
	switch noticeType {
	case NoticeTypeAnnouncement:
		return noticeCategoryCodeAnnouncement
	default:
		return noticeCategoryCodeNotice
	}
}
