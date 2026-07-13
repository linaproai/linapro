// This file tests email channel config parsing and unavailable delivery errors.

package notify

import (
	"context"
	"testing"

	"lina-core/internal/model/entity"
	"lina-core/pkg/plugin/capability/notifycap"
)

func TestParseEmailChannelAccountID(t *testing.T) {
	id, err := parseEmailChannelAccountID("")
	if err != nil || id != 0 {
		t.Fatalf("empty config: id=%d err=%v", id, err)
	}
	id, err = parseEmailChannelAccountID(`{"accountId":42}`)
	if err != nil || id != 42 {
		t.Fatalf("account id: id=%d err=%v", id, err)
	}
}

func TestSendEmailUnavailableWithoutBridge(t *testing.T) {
	notifycap.ResetEmailDeliveryForTest()
	svc := &serviceImpl{}
	_, err := svc.sendEmail(context.Background(), &entity.SysNotifyChannel{
		ChannelKey:  "email",
		ChannelType: ChannelTypeEmail.String(),
		ConfigJson:  "{}",
	}, SendInput{
		Title:            "t",
		Content:          "c",
		RecipientUserIDs: []int64{1},
	})
	if err == nil {
		t.Fatal("expected unavailable error")
	}
}
