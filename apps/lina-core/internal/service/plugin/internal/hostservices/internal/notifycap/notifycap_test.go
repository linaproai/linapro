// This file verifies notification capability DTO conversion inside notifycap.

package notifycap

import (
	"context"
	"testing"

	notifysvc "lina-core/internal/service/notify"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
)

// TestSendConvertsNoticeDTOs verifies source-plugin notification DTOs are
// converted to host notify DTOs inside the notification capability component.
func TestSendConvertsNoticeDTOs(t *testing.T) {
	ctx := context.Background()
	publisher := &fakeNotifyPublisher{}
	svc := New(publisher)

	output, err := svc.Send(ctx, capmodel.CapabilityContext{}, capabilitynotifycap.SendInput{
		SourceType:   capabilitynotifycap.SourceTypeNotice,
		SourceID:     "1001",
		Title:        "Release",
		Content:      "Published",
		Category:     capabilitynotifycap.CategoryCodeOther,
		SenderUserID: 42,
	})
	if err != nil {
		t.Fatalf("send notice publication: %v", err)
	}
	if output == nil || output.MessageID != "9001" || output.DeliveryCount != 3 {
		t.Fatalf("unexpected notify output: %#v", output)
	}
	if publisher.noticeInput.NoticeID != 1001 ||
		publisher.noticeInput.CategoryCode != notifysvc.CategoryCodeOther ||
		publisher.noticeInput.SenderUserID != 42 {
		t.Fatalf("expected host notify input to be recorded, got %#v", publisher.noticeInput)
	}

	if err = svc.DeleteBySource(ctx, capmodel.CapabilityContext{}, capabilitynotifycap.SourceTypeNotice, []string{"1001"}); err != nil {
		t.Fatalf("delete by source: %v", err)
	}
	if publisher.deletedSourceType != notifysvc.SourceTypeNotice ||
		len(publisher.deletedSourceIDs) != 1 ||
		publisher.deletedSourceIDs[0] != "1001" {
		t.Fatalf("expected host notify delete input to be recorded, got %q %#v", publisher.deletedSourceType, publisher.deletedSourceIDs)
	}
}

// fakeNotifyPublisher records notify DTOs passed to the host notify boundary.
type fakeNotifyPublisher struct {
	noticeInput       notifysvc.NoticePublishInput
	sendInput         notifysvc.SendInput
	deletedSourceType notifysvc.SourceType
	deletedSourceIDs  []string
}

// Send records one host notify send input.
func (f *fakeNotifyPublisher) Send(_ context.Context, in notifysvc.SendInput) (*notifysvc.SendOutput, error) {
	f.sendInput = in
	return &notifysvc.SendOutput{MessageID: 9002, DeliveryCount: len(in.RecipientUserIDs)}, nil
}

// SendNoticePublication records one host notify publication input.
func (f *fakeNotifyPublisher) SendNoticePublication(
	_ context.Context,
	in notifysvc.NoticePublishInput,
) (*notifysvc.SendOutput, error) {
	f.noticeInput = in
	return &notifysvc.SendOutput{MessageID: 9001, DeliveryCount: 3}, nil
}

// DeleteBySource records one host notify delete request.
func (f *fakeNotifyPublisher) DeleteBySource(_ context.Context, sourceType notifysvc.SourceType, sourceIDs []string) error {
	f.deletedSourceType = sourceType
	f.deletedSourceIDs = append([]string(nil), sourceIDs...)
	return nil
}
