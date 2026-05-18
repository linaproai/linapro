// This file verifies runtime session-timeout validation during request access.

package session

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
	_ "lina-core/pkg/dbdriver"

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
	exists, err := store.TouchOrValidate(ctx, 0, tokenID, time.Hour)
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

// TestIsSessionInactiveUsesTimeout verifies expiration is checked before a
// persistent session row can be treated as valid.
func TestIsSessionInactiveUsesTimeout(t *testing.T) {
	now := gtime.Now()
	stored := &entity.SysOnlineSession{LastActiveTime: now.Add(-2 * time.Hour)}
	if !isSessionInactive(stored, now, time.Hour) {
		t.Fatal("expected stale session row to be inactive")
	}
	if isSessionInactive(stored, now, 0) {
		t.Fatal("expected disabled timeout to keep the session active")
	}
}

// TestTouchOrValidateRefreshesActiveSession verifies valid sessions keep their
// record and refresh the last-active timestamp.
func TestTouchOrValidateRefreshesActiveSession(t *testing.T) {
	ctx := context.Background()
	tokenID := fmt.Sprintf("session-active-%d", time.Now().UnixNano())
	lastActive := gtime.Now().Add(-2 * sessionLastActiveUpdateWindow).Truncate(time.Second)

	insertSessionRecord(t, ctx, tokenID, lastActive)

	store := NewDBStore()
	exists, err := store.TouchOrValidate(ctx, 0, tokenID, time.Hour)
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

// TestSessionRecordSurvivesStoreRecreation verifies valid online-session rows
// are retained across process-local store recreation and remain usable.
func TestSessionRecordSurvivesStoreRecreation(t *testing.T) {
	ctx := context.Background()
	tokenID := fmt.Sprintf("session-restart-%d", time.Now().UnixNano())
	lastActive := gtime.Now().Add(-2 * sessionLastActiveUpdateWindow).Truncate(time.Second)

	insertSessionRecord(t, ctx, tokenID, lastActive)

	firstStore := NewDBStore()
	stored, err := firstStore.Get(ctx, tokenID)
	if err != nil {
		t.Fatalf("get session before store recreation: %v", err)
	}
	if stored == nil {
		t.Fatal("expected session before store recreation")
	}

	secondStore := NewDBStore()
	exists, err := secondStore.TouchOrValidate(ctx, 0, tokenID, time.Hour)
	if err != nil {
		t.Fatalf("touch session after store recreation: %v", err)
	}
	if !exists {
		t.Fatal("expected unexpired session to survive store recreation")
	}
}

// TestTouchOrValidateRejectsTenantMismatch verifies token identity is global
// while request validation still enforces the expected tenant ownership.
func TestTouchOrValidateRejectsTenantMismatch(t *testing.T) {
	ctx := context.Background()
	tokenID := fmt.Sprintf("session-tenant-mismatch-%d", time.Now().UnixNano())
	lastActive := gtime.Now().Add(-2 * sessionLastActiveUpdateWindow).Truncate(time.Second)

	insertTenantSessionRecord(t, ctx, 22, tokenID, lastActive)

	store := NewDBStore()
	stored, err := store.Get(ctx, tokenID)
	if err != nil {
		t.Fatalf("get tenant session: %v", err)
	}
	if stored == nil || stored.TenantId != 22 {
		t.Fatalf("expected tenant 22 session, got %#v", stored)
	}
	exists, err := store.TouchOrValidate(ctx, 11, tokenID, time.Hour)
	if err != nil {
		t.Fatalf("touch mismatched tenant session: %v", err)
	}
	if exists {
		t.Fatal("expected mismatched tenant session to be rejected")
	}
	exists, err = store.TouchOrValidate(ctx, 22, tokenID, time.Hour)
	if err != nil {
		t.Fatalf("touch matching tenant session: %v", err)
	}
	if !exists {
		t.Fatal("expected matching tenant session to remain valid")
	}
}

// TestTouchOrValidateSkipsRecentActiveSessionUpdate verifies recent activity
// remains valid without writing another last-active timestamp.
func TestTouchOrValidateSkipsRecentActiveSessionUpdate(t *testing.T) {
	ctx := context.Background()
	tokenID := fmt.Sprintf("session-recent-%d", time.Now().UnixNano())
	lastActive := gtime.Now().Truncate(time.Second)

	insertSessionRecord(t, ctx, tokenID, lastActive)

	store := NewDBStore()
	exists, err := store.TouchOrValidate(ctx, 0, tokenID, time.Hour)
	if err != nil {
		t.Fatalf("touch recent session: %v", err)
	}
	if !exists {
		t.Fatal("expected recent session to remain valid")
	}

	var stored *entity.SysOnlineSession
	err = dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenID}).
		Scan(&stored)
	if err != nil {
		t.Fatalf("query recent session after validation: %v", err)
	}
	if stored == nil || stored.LastActiveTime == nil {
		t.Fatal("expected recent session record to remain after validation")
	}
	if stored.LastActiveTime.String() != lastActive.String() {
		t.Fatalf("expected recent last active time to remain unchanged, got %v want %v", stored.LastActiveTime, lastActive)
	}
}

// insertSessionRecord inserts one online-session row used by validation tests
// and registers cleanup automatically.
func insertSessionRecord(t *testing.T, ctx context.Context, tokenID string, lastActive *gtime.Time) {
	t.Helper()
	insertTenantSessionRecord(t, ctx, 0, tokenID, lastActive)
}

// insertTenantSessionRecord inserts one online-session row for a specific
// tenant and registers cleanup automatically.
func insertTenantSessionRecord(t *testing.T, ctx context.Context, tenantID int, tokenID string, lastActive *gtime.Time) {
	t.Helper()

	_, err := dao.SysOnlineSession.Ctx(ctx).Data(do.SysOnlineSession{
		TokenId:        tokenID,
		TenantId:       tenantID,
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
