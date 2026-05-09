// This file verifies scheduled-job execution-log queries against SQLite.

package jobmgmt

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/dialect"
)

// sqliteListLogsChildEnv marks the isolated child process that owns SQLite global config.
const sqliteListLogsChildEnv = "LINA_SQLITE_JOB_LOG_LIST_CHILD"

// TestSQLiteListLogsAcceptsScopedExists verifies execution-log list queries do
// not produce SQLite-invalid EXISTS double parentheses.
func TestSQLiteListLogsAcceptsScopedExists(t *testing.T) {
	if os.Getenv(sqliteListLogsChildEnv) == "1" {
		t.Skip("parent test only launches the isolated SQLite child process")
	}

	dbPath := filepath.Join(t.TempDir(), "linapro-job-log-list.db")
	cmd := exec.Command(os.Args[0], "-test.run=^TestSQLiteListLogsAcceptsScopedExistsChild$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(),
		sqliteListLogsChildEnv+"=1",
		"LINA_SQLITE_JOB_LOG_LIST_DB="+dbPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("SQLite job log list child test failed: %v\n%s", err, string(output))
	}
}

// TestSQLiteListLogsAcceptsScopedExistsChild runs the actual SQLite regression
// check inside an isolated process because GoFrame database config is global.
func TestSQLiteListLogsAcceptsScopedExistsChild(t *testing.T) {
	if os.Getenv(sqliteListLogsChildEnv) != "1" {
		t.Skip("SQLite job log list child test is executed by TestSQLiteListLogsAcceptsScopedExists")
	}

	ctx := context.Background()
	link := sqliteJobLogListLinkFromEnv(t)
	setupSQLiteJobLogListDatabase(t, ctx, link)

	adminID := sqliteJobLogListAdminID(t, ctx)
	jobID := insertLogCleanupTestJob(t, ctx)
	logID := insertLogCleanupTestLog(t, ctx, jobID, "sqlite-list")

	svc := newTestService(t)
	svc.bizCtxSvc = jobmgmtStaticBizCtx{ctx: &model.Context{UserId: adminID}}
	out, err := svc.ListLogs(ctx, ListLogsInput{PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list SQLite execution logs: %v", err)
	}
	if out == nil || out.Total < 1 {
		t.Fatalf("expected SQLite execution-log list to return rows, got %#v", out)
	}
	if _, ok := jobScopeLogIDSet(out.List)[logID]; !ok {
		t.Fatalf("expected SQLite execution-log list to include log %d, got %#v", logID, jobScopeLogIDSet(out.List))
	}
}

// sqliteJobLogListLinkFromEnv builds the SQLite link used by the child test.
func sqliteJobLogListLinkFromEnv(t *testing.T) string {
	t.Helper()

	dbPath := os.Getenv("LINA_SQLITE_JOB_LOG_LIST_DB")
	if dbPath == "" {
		t.Fatal("LINA_SQLITE_JOB_LOG_LIST_DB must be set")
	}
	return "sqlite::@file(" + dbPath + ")"
}

// setupSQLiteJobLogListDatabase points GoFrame at a temporary SQLite database
// and initializes the host schema in the same order as regular init.
func setupSQLiteJobLogListDatabase(t *testing.T, ctx context.Context, link string) {
	t.Helper()

	dbDialect, err := dialect.From(link)
	if err != nil {
		t.Fatalf("resolve SQLite dialect: %v", err)
	}
	if err = dbDialect.PrepareDatabase(ctx, link, true); err != nil {
		t.Fatalf("prepare SQLite job-log database: %v", err)
	}
	if err = gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: link}},
	}); err != nil {
		t.Fatalf("configure SQLite job-log database: %v", err)
	}
	adapter, err := gcfg.NewAdapterContent(sqliteJobLogListConfig(link))
	if err != nil {
		t.Fatalf("create SQLite job-log config adapter: %v", err)
	}
	g.Cfg().SetAdapter(adapter)

	for _, asset := range sqliteJobLogListHostSQLAssets(t) {
		translated, translateErr := dbDialect.TranslateDDL(ctx, asset.path, asset.content)
		if translateErr != nil {
			t.Fatalf("translate SQLite host SQL %s: %v", asset.path, translateErr)
		}
		for index, statement := range dialect.SplitSQLStatements(translated) {
			if _, err = g.DB().Exec(ctx, statement); err != nil {
				t.Fatalf("execute SQLite host SQL %s statement %d: %v\n%s", asset.path, index+1, err, statement)
			}
		}
	}
	t.Cleanup(func() {
		if closeErr := g.DB().Close(ctx); closeErr != nil {
			t.Fatalf("close SQLite job-log database: %v", closeErr)
		}
	})
}

// sqliteJobLogListConfig returns the minimal runtime configuration needed by
// role, i18n, and data-scope services during the SQLite regression test.
func sqliteJobLogListConfig(link string) string {
	return `database:
  default:
    link: "` + link + `"
cluster:
  enabled: false
jwt:
  secret: "sqlite-job-log-list-test-secret"
  expire: 24h
i18n:
  default: zh-CN
  enabled: true
  locales:
    - locale: en-US
      nativeName: English
    - locale: zh-CN
      nativeName: Simplified Chinese
`
}

// sqliteJobLogListSQLAsset stores one ordered host SQL file for the test setup.
type sqliteJobLogListSQLAsset struct {
	path    string
	content string
}

// sqliteJobLogListHostSQLAssets reads host install SQL files from the current
// source tree in lexical order.
func sqliteJobLogListHostSQLAssets(t *testing.T) []sqliteJobLogListSQLAsset {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	sqlDir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "manifest", "sql")
	entries, err := os.ReadDir(sqlDir)
	if err != nil {
		t.Fatalf("read host SQL directory: %v", err)
	}
	assets := make([]sqliteJobLogListSQLAsset, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		assetPath := filepath.Join(sqlDir, entry.Name())
		content, readErr := os.ReadFile(assetPath)
		if readErr != nil {
			t.Fatalf("read host SQL asset %s: %v", assetPath, readErr)
		}
		assets = append(assets, sqliteJobLogListSQLAsset{path: assetPath, content: string(content)})
	}
	sort.SliceStable(assets, func(i int, j int) bool {
		return assets[i].path < assets[j].path
	})
	return assets
}

// sqliteJobLogListAdminID returns the seeded administrator user identifier.
func sqliteJobLogListAdminID(t *testing.T, ctx context.Context) int {
	t.Helper()

	var user *entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Username: "admin"}).
		Scan(&user); err != nil {
		t.Fatalf("query SQLite admin user: %v", err)
	}
	if user == nil || user.Id <= 0 {
		t.Fatal("expected SQLite admin user seed to exist")
	}
	return user.Id
}
