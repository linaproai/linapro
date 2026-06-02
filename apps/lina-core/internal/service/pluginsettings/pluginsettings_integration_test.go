// This file verifies database-backed regression scenarios for the host
// plugin settings service. The scenarios pin down two contracts that are
// easy to break and impossible to validate through the pure helpers in
// pluginsettings_test.go:
//
//   - SetString("") followed by SetString("new") must produce a row that is
//     visible to GetString and List. The historical bug here was a soft
//     delete that left a ghost row under the (tenant_id, key) unique index,
//     which made later upserts update the invisible row in place.
//   - Soft-deleted rows that already exist (created before this fix landed,
//     or by parallel writers that ran against the previous implementation)
//     must be recovered to the visible state when the same key is upserted
//     again, so the system self-heals without manual SQL.
//
// Execution requirements:
//
//   - A PostgreSQL instance reachable through the GoFrame database link
//     declared by the active config (defaults to localhost:5432, database
//     "linapro"). Other integration tests in this repository assume the
//     same setup; running `make init` once prepares the schema.
//   - The tests register their own row per run via uniqueIntegrationPluginID
//     and clean up via t.Cleanup so parallel runs do not collide.
//   - These tests are not part of the pure pluginsettings_test.go file
//     because that file is intentionally DB-free; the comment there
//     reserves DB-backed scenarios for this file.

package pluginsettings

import (
	"context"
	"fmt"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"

	// Blank-import the host driver registry so the GoFrame pgsql driver
	// is registered before dao.SysConfig touches g.DB(). Other integration
	// tests in the repository rely on the same registration side effect,
	// which normally comes from main.go but is absent under `go test`.
	_ "lina-core/pkg/dbdriver"
)

// TestSetStringRecoversAfterClearAndRewrite verifies the public contract
// "clear a setting, then write a new value, then read the new value back"
// stays observable. This is the user-visible scenario the reviewer flagged:
// an admin clears backendRedirects, configures a new JSON later, and
// expects /plugin/<id>/settings GET to return the new JSON.
func TestSetStringRecoversAfterClearAndRewrite(t *testing.T) {
	ctx := context.Background()
	svc := New()
	pluginID := uniqueIntegrationPluginID(t)
	key := "recoverAfterClearAndRewrite"
	t.Cleanup(func() { cleanupIntegrationKey(t, ctx, pluginID, key) })

	if err := svc.SetString(ctx, pluginID, key, "initial"); err != nil {
		t.Fatalf("seed initial value: %v", err)
	}
	if value, err := svc.GetString(ctx, pluginID, key, ""); err != nil || value != "initial" {
		t.Fatalf("initial GetString = %q err=%v, want %q", value, err, "initial")
	}

	if err := svc.SetString(ctx, pluginID, key, ""); err != nil {
		t.Fatalf("clear value: %v", err)
	}
	if value, err := svc.GetString(ctx, pluginID, key, "fallback"); err != nil || value != "fallback" {
		t.Fatalf("post-clear GetString = %q err=%v, want fallback to default", value, err)
	}

	if err := svc.SetString(ctx, pluginID, key, "next"); err != nil {
		t.Fatalf("rewrite value: %v", err)
	}
	value, err := svc.GetString(ctx, pluginID, key, "")
	if err != nil {
		t.Fatalf("post-rewrite GetString: %v", err)
	}
	if value != "next" {
		t.Fatalf("post-rewrite GetString = %q, want %q (clear-then-rewrite must produce a visible row)", value, "next")
	}

	listed, err := svc.List(ctx, pluginID)
	if err != nil {
		t.Fatalf("List after rewrite: %v", err)
	}
	if listed[key] != "next" {
		t.Fatalf("List after rewrite[%s] = %q, want %q", key, listed[key], "next")
	}
}

// TestUpsertValueRecoversSoftDeletedRow verifies the upsert path heals a
// pre-existing soft-deleted row by clearing deleted_at on conflict. This
// scenario simulates legacy data written before the fix shipped (where the
// previous Delete path created soft-deleted rows) or a parallel writer
// that still uses the legacy semantics. Without this recovery, the row
// would stay invisible to GetString/List after the next SetString call.
func TestUpsertValueRecoversSoftDeletedRow(t *testing.T) {
	ctx := context.Background()
	svc := New()
	pluginID := uniqueIntegrationPluginID(t)
	key := "recoverFromLegacySoftDelete"
	fullKey := pluginID + keySeparator + key
	t.Cleanup(func() { cleanupIntegrationKey(t, ctx, pluginID, key) })

	if err := svc.SetString(ctx, pluginID, key, "initial"); err != nil {
		t.Fatalf("seed initial value: %v", err)
	}

	// Simulate the legacy clear path that soft-deletes the row instead of
	// physically removing it. After this step the row is invisible to
	// GetString because GoFrame applies its automatic
	// "deleted_at IS NULL" filter to every dao.SysConfig query.
	if _, err := dao.SysConfig.Ctx(ctx).
		Where(dao.SysConfig.Columns().TenantId, platformTenantID).
		Where(dao.SysConfig.Columns().Key, fullKey).
		Delete(); err != nil {
		t.Fatalf("soft-delete via legacy path: %v", err)
	}
	if value, err := svc.GetString(ctx, pluginID, key, "fallback"); err != nil || value != "fallback" {
		t.Fatalf("post-soft-delete GetString = %q err=%v, want fallback", value, err)
	}

	if err := svc.SetString(ctx, pluginID, key, "recovered"); err != nil {
		t.Fatalf("rewrite over soft-deleted row: %v", err)
	}

	value, err := svc.GetString(ctx, pluginID, key, "")
	if err != nil {
		t.Fatalf("post-recovery GetString: %v", err)
	}
	if value != "recovered" {
		t.Fatalf("post-recovery GetString = %q, want %q (upsert must reset deleted_at)", value, "recovered")
	}

	// Defense in depth: verify the underlying row's deleted_at is actually
	// NULL. A future change that drops deleted_at from OnDuplicate would
	// still pass the GetString check above only when no soft-deleted row
	// existed; this assertion catches the regression directly.
	var row *entity.SysConfig
	if err := dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(dao.SysConfig.Columns().TenantId, platformTenantID).
		Where(dao.SysConfig.Columns().Key, fullKey).
		Scan(&row); err != nil {
		t.Fatalf("scan recovered row: %v", err)
	}
	if row == nil {
		t.Fatalf("recovered row should exist for key %s", fullKey)
	}
	if row.DeletedAt != nil {
		t.Fatalf("recovered row deleted_at = %v, want nil", row.DeletedAt)
	}
	if row.Value != "recovered" {
		t.Fatalf("recovered row value = %q, want %q", row.Value, "recovered")
	}
}

// TestSetStringClearRemovesRowPhysically verifies the clear path actually
// removes the row from sys_config instead of soft-deleting it, so future
// upserts do not have to dance with a ghost row under the unique index.
func TestSetStringClearRemovesRowPhysically(t *testing.T) {
	ctx := context.Background()
	svc := New()
	pluginID := uniqueIntegrationPluginID(t)
	key := "clearRemovesRowPhysically"
	fullKey := pluginID + keySeparator + key
	t.Cleanup(func() { cleanupIntegrationKey(t, ctx, pluginID, key) })

	if err := svc.SetString(ctx, pluginID, key, "value"); err != nil {
		t.Fatalf("seed value: %v", err)
	}
	if err := svc.SetString(ctx, pluginID, key, ""); err != nil {
		t.Fatalf("clear value: %v", err)
	}

	count, err := dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(dao.SysConfig.Columns().TenantId, platformTenantID).
		Where(dao.SysConfig.Columns().Key, fullKey).
		Count()
	if err != nil {
		t.Fatalf("count after clear: %v", err)
	}
	if count != 0 {
		t.Fatalf("post-clear row count = %d, want 0 (Unscoped delete must remove the row)", count)
	}
}

// uniqueIntegrationPluginID returns a per-test pluginID so parallel runs do
// not collide on the same (tenant_id, key) unique index entry.
func uniqueIntegrationPluginID(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("linapro-pluginsettings-itest-%d", time.Now().UnixNano())
}

// cleanupIntegrationKey removes the row used by one integration test even
// if the test failed before reaching its own cleanup path.
func cleanupIntegrationKey(t *testing.T, ctx context.Context, pluginID string, key string) {
	t.Helper()
	fullKey := pluginID + keySeparator + key
	_, err := dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(dao.SysConfig.Columns().TenantId, platformTenantID).
		Where(dao.SysConfig.Columns().Key, fullKey).
		Delete()
	if err != nil {
		t.Fatalf("cleanup integration row %s: %v", fullKey, err)
	}
}
