// This file covers dynamic-route-specific session validation behaviors that
// are easy to regress during runtime auth changes.

package runtime

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
)

func TestTouchDynamicRouteSessionKeepsExistingSessionWhenTimestampDoesNotChange(t *testing.T) {
	var (
		ctx     = context.Background()
		service = &serviceImpl{}
		tokenID = fmt.Sprintf("plugin-dynamic-route-session-%d", time.Now().UnixNano())
	)

	if _, err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete stale online session %s: %v", tokenID, err)
	}
	defer func() {
		if _, err := dao.SysOnlineSession.Ctx(ctx).
			Where(do.SysOnlineSession{TokenId: tokenID}).
			Delete(); err != nil {
			t.Fatalf("failed to cleanup online session %s: %v", tokenID, err)
		}
	}()

	currentSecond := waitForFreshSecond(t)
	_, err := dao.SysOnlineSession.Ctx(ctx).Data(do.SysOnlineSession{
		TokenId:        tokenID,
		UserId:         1,
		Username:       "admin",
		DeptName:       "系统管理",
		Ip:             "127.0.0.1",
		Browser:        "go-test",
		Os:             "darwin",
		LoginTime:      currentSecond,
		LastActiveTime: currentSecond,
	}).Insert()
	if err != nil {
		t.Fatalf("expected test session insert to succeed, got error: %v", err)
	}

	exists, err := service.touchDynamicRouteSession(ctx, tokenID)
	if err != nil {
		t.Fatalf("expected first session touch to succeed, got error: %v", err)
	}
	if !exists {
		t.Fatal("expected first session touch to keep the session active")
	}

	// Touch the same session again within the same second to emulate the dynamic
	// route request arriving immediately after login or another protected request.
	exists, err = service.touchDynamicRouteSession(ctx, tokenID)
	if err != nil {
		t.Fatalf("expected second session touch to succeed, got error: %v", err)
	}
	if !exists {
		t.Fatal("expected existing session to remain active when DATETIME precision keeps the same second")
	}
}

func waitForFreshSecond(t *testing.T) *gtime.Time {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		now := time.Now()
		if now.Nanosecond() < int((100 * time.Millisecond).Nanoseconds()) {
			return gtime.NewFromTime(now.Truncate(time.Second))
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("failed to align test to the beginning of a second")
	return nil
}
