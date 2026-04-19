// This file verifies runtime session-timeout validation during request access.

package session

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// TestTouchOrValidateRejectsExpiredSession verifies expired online sessions are
// removed during validation.
func TestTouchOrValidateRejectsExpiredSession(t *testing.T) {
	ctx := context.Background()
	tokenID := fmt.Sprintf("session-expired-%d", time.Now().UnixNano())

	insertSessionRecord(t, ctx, tokenID, gtime.Now().Add(-2*time.Hour))

	store := NewDBStore()
	exists, err := store.TouchOrValidate(ctx, tokenID, time.Hour)
	if err != nil {
		t.Fatalf("touch expired session: %v", err)
	}
	if exists {
		t.Fatal("expected expired session to be rejected")
	}

	var stored *entity.SysOnlineSession
	err = dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenID}).
		Scan(&stored)
	if err != nil {
		t.Fatalf("query expired session after validation: %v", err)
	}
	if stored != nil {
		t.Fatal("expected expired session to be deleted")
	}
}

// TestTouchOrValidateRefreshesActiveSession verifies valid sessions keep their
// record and refresh the last-active timestamp.
func TestTouchOrValidateRefreshesActiveSession(t *testing.T) {
	ctx := context.Background()
	tokenID := fmt.Sprintf("session-active-%d", time.Now().UnixNano())
	lastActive := gtime.Now().Add(-10 * time.Minute)

	insertSessionRecord(t, ctx, tokenID, lastActive)

	store := NewDBStore()
	exists, err := store.TouchOrValidate(ctx, tokenID, time.Hour)
	if err != nil {
		t.Fatalf("touch active session: %v", err)
	}
	if !exists {
		t.Fatal("expected active session to remain valid")
	}

	var stored *entity.SysOnlineSession
	err = dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenID}).
		Scan(&stored)
	if err != nil {
		t.Fatalf("query active session after validation: %v", err)
	}
	if stored == nil || stored.LastActiveTime == nil {
		t.Fatal("expected active session record to remain after validation")
	}
	if !stored.LastActiveTime.After(lastActive) {
		t.Fatalf("expected last active time to be refreshed, got %v", stored.LastActiveTime)
	}
}

// insertSessionRecord inserts one online-session row used by validation tests
// and registers cleanup automatically.
func insertSessionRecord(t *testing.T, ctx context.Context, tokenID string, lastActive *gtime.Time) {
	t.Helper()

	_, err := dao.SysOnlineSession.Ctx(ctx).Data(do.SysOnlineSession{
		TokenId:        tokenID,
		UserId:         1,
		Username:       "admin",
		DeptName:       "系统管理部",
		Ip:             "127.0.0.1",
		Browser:        "test",
		Os:             "darwin",
		LoginTime:      lastActive,
		LastActiveTime: lastActive,
	}).Insert()
	if err != nil {
		t.Fatalf("insert session record %s: %v", tokenID, err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysOnlineSession.Ctx(ctx).Where(do.SysOnlineSession{TokenId: tokenID}).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup session record %s: %v", tokenID, cleanupErr)
		}
	})
}
