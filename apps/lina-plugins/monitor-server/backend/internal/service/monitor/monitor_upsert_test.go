// This file verifies monitor-server snapshot persistence across SQL dialects.

package monitor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/dialect"
	"lina-plugin-monitor-server/backend/internal/dao"
)

// TestUpsertMonitorSnapshotWorksOnSQLite verifies SQLite runtime persistence
// uses explicit Save conflict columns for monitor snapshots.
func TestUpsertMonitorSnapshotWorksOnSQLite(t *testing.T) {
	ctx := context.Background()
	setupSQLiteMonitorServerDB(t, ctx)

	const (
		nodeName   = "unit-node"
		nodeIP     = "10.0.0.10"
		firstData  = `{"sample":1}`
		secondData = `{"sample":2}`
	)

	if err := upsertMonitorSnapshot(ctx, nodeName, nodeIP, firstData); err != nil {
		t.Fatalf("insert SQLite monitor snapshot failed: %v", err)
	}
	if err := upsertMonitorSnapshot(ctx, nodeName, nodeIP, secondData); err != nil {
		t.Fatalf("update SQLite monitor snapshot failed: %v", err)
	}

	count, err := dao.Server.Ctx(ctx).Count()
	if err != nil {
		t.Fatalf("count SQLite monitor snapshots failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one snapshot row after duplicate node upsert, got %d", count)
	}

	var row *serverMonitorRecord
	if err = dao.Server.Ctx(ctx).Where(colNodeName, nodeName).Scan(&row); err != nil {
		t.Fatalf("read SQLite monitor snapshot failed: %v", err)
	}
	if row == nil {
		t.Fatal("expected SQLite monitor snapshot row to exist")
	}
	if row.Data != secondData {
		t.Fatalf("expected latest monitor data %s, got %s", secondData, row.Data)
	}
}

// TestGetDBInfoReturnsSQLiteVersion verifies monitor-server database
// diagnostics return a non-empty SQLite version label instead of silently
// swallowing the MySQL-only VERSION() failure.
func TestGetDBInfoReturnsSQLiteVersion(t *testing.T) {
	ctx := context.Background()
	setupSQLiteMonitorServerDB(t, ctx)

	info := New().GetDBInfo(ctx)
	if info == nil {
		t.Fatal("expected SQLite DB info to be returned")
	}
	if !strings.HasPrefix(info.Version, "SQLite ") {
		t.Fatalf("expected SQLite database version label, got %q", info.Version)
	}
	if strings.TrimSpace(strings.TrimPrefix(info.Version, "SQLite ")) == "" {
		t.Fatalf("expected SQLite database version number to be non-empty, got %q", info.Version)
	}
}

// setupSQLiteMonitorServerDB points the generated DAO at a temporary SQLite
// database and creates the monitor-server table.
func setupSQLiteMonitorServerDB(t *testing.T, ctx context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	dbPath := filepath.Join(t.TempDir(), "monitor-server.db")
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: "sqlite::@file(" + dbPath + ")"}},
	}); err != nil {
		t.Fatalf("configure SQLite monitor database failed: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if closeErr := db.Close(ctx); closeErr != nil {
			t.Errorf("close SQLite monitor database failed: %v", closeErr)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore GoFrame database config failed: %v", err)
		}
	})

	dbDialect, err := dialect.From("sqlite::@file(" + dbPath + ")")
	if err != nil {
		t.Fatalf("resolve SQLite dialect failed: %v", err)
	}
	sqlPath := filepath.Join("..", "..", "..", "..", "manifest", "sql", "001-monitor-server-schema.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("read monitor-server schema SQL failed: %v", err)
	}
	translated, err := dbDialect.TranslateDDL(ctx, sqlPath, string(content))
	if err != nil {
		t.Fatalf("translate monitor-server schema SQL failed: %v", err)
	}
	for _, statement := range dialect.SplitSQLStatements(translated) {
		if _, err = db.Exec(ctx, statement); err != nil {
			t.Fatalf("execute monitor-server schema SQL failed: %v\nSQL:\n%s", err, statement)
		}
	}
}
