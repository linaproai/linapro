// This file verifies notification capability DTO conversion inside notifycap.

package notify

import (
	"context"
	"strconv"
	"testing"

	usermsgv1 "lina-core/api/usermsg/v1"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
)

// TestSendConvertsNoticeDTOs verifies source-plugin notification DTOs are
// converted to host notify DTOs inside the notification capability component.
func TestSendConvertsNoticeDTOs(t *testing.T) {
	ctx := bizctxcap.WithCurrentContext(context.Background(), bizctxcap.CurrentContext{
		UserID:   42,
		TenantID: 7,
	})
	publisher := &fakeNotifyPublisher{}
	svc := NewCapabilityAdapter(publisher, nil)

	output, err := svc.Send(ctx, capabilitynotifycap.SendInput{
		SourceType:   usermsgv1.SourceTypeNotice,
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
		publisher.noticeInput.CategoryCode != CategoryCodeOther ||
		publisher.noticeInput.SenderUserID != 42 {
		t.Fatalf("expected host notify input to be recorded, got %#v", publisher.noticeInput)
	}

	if err = svc.DeleteBySource(ctx, usermsgv1.SourceTypeNotice, []string{"1001"}); err != nil {
		t.Fatalf("delete by source: %v", err)
	}
	if publisher.deletedSourceType != usermsgv1.SourceTypeNotice ||
		len(publisher.deletedSourceIDs) != 1 ||
		publisher.deletedSourceIDs[0] != "1001" {
		t.Fatalf("expected host notify delete input to be recorded, got %q %#v", publisher.deletedSourceType, publisher.deletedSourceIDs)
	}
}

// TestBatchGetBySourceRejectsSourceIDLimit verifies source-based notification
// reads have a bounded input set before any storage query is built.
func TestBatchGetBySourceRejectsSourceIDLimit(t *testing.T) {
	ids := make([]string, capabilitynotifycap.MaxBatchGetBySourceIDs+1)
	for i := range ids {
		ids[i] = "source-" + strconv.Itoa(i)
	}
	ctx := bizctxcap.WithCurrentContext(context.Background(), bizctxcap.CurrentContext{
		UserID:   1,
		TenantID: 7,
	})
	_, err := NewCapabilityAdapter(nil, nil).BatchGetBySource(ctx, capabilitynotifycap.BatchGetBySourceInput{
		SourceType: usermsgv1.SourceTypePlugin,
		SourceIDs:  ids,
	})
	if !bizerr.Is(err, capmodel.CodeCapabilityLimitExceeded) {
		t.Fatalf("expected limit error, got %v", err)
	}
}

// fakeNotifyPublisher records notify DTOs passed to the host notify boundary.
type fakeNotifyPublisher struct {
	Service
	noticeInput       NoticePublishInput
	sendInput         SendInput
	deletedSourceType SourceType
	deletedSourceIDs  []string
}

// Send records one host notification send input.
func (f *fakeNotifyPublisher) Send(_ context.Context, in SendInput) (*SendOutput, error) {
	f.sendInput = in
	return &SendOutput{MessageID: 9002, DeliveryCount: len(in.RecipientUserIDs)}, nil
}

// SendNoticePublication records one host notify publication input.
func (f *fakeNotifyPublisher) SendNoticePublication(
	_ context.Context,
	in NoticePublishInput,
) (*SendOutput, error) {
	f.noticeInput = in
	return &SendOutput{MessageID: 9001, DeliveryCount: 3}, nil
}

// DeleteBySource records one host notify delete request.
func (f *fakeNotifyPublisher) DeleteBySource(_ context.Context, sourceType SourceType, sourceIDs []string) error {
	f.deletedSourceType = sourceType
	f.deletedSourceIDs = append([]string(nil), sourceIDs...)
	return nil
}
