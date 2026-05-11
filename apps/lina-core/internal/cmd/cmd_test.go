// This file verifies shared command helpers for explicit confirmations, SQL
// asset source selection, and SQL execution behavior.

package cmd

import (
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/glog"
	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/service/startupstats"
	"lina-core/pkg/logger"
)

// TestRequireCommandConfirmation verifies sensitive command confirmation tokens
// are enforced for init and mock operations.
func TestRequireCommandConfirmation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		commandName    string
		confirmValue   string
		wantErr        bool
		wantSubstrings []string
	}{
		{
			name:         "init accepts matching confirmation",
			commandName:  initCommandName,
			confirmValue: initCommandName,
		},
		{
			name:         "mock accepts matching confirmation",
			commandName:  mockCommandName,
			confirmValue: mockCommandName,
		},
		{
			name:         "init rejects missing confirmation",
			commandName:  initCommandName,
			confirmValue: "",
			wantErr:      true,
			wantSubstrings: []string{
				"command init performs sensitive upgrade or database operations",
				makeConfirmationExample(initCommandName),
				goRunConfirmationExample(initCommandName),
			},
		},
		{
			name:         "mock rejects wrong confirmation",
			commandName:  mockCommandName,
			confirmValue: initCommandName,
			wantErr:      true,
			wantSubstrings: []string{
				"command mock performs sensitive upgrade or database operations",
				makeConfirmationExample(mockCommandName),
				goRunConfirmationExample(mockCommandName),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := requireCommandConfirmation(tt.commandName, tt.confirmValue)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for command %q", tt.commandName)
				}
				for _, substring := range tt.wantSubstrings {
					if !strings.Contains(err.Error(), substring) {
						t.Fatalf("expected error %q to contain %q", err.Error(), substring)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// TestCommandPackageHasNoHanText verifies CLI diagnostics in this package stay
// as English developer-facing source text.
func TestCommandPackageHasNoHanText(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read command package directory: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		content, readErr := os.ReadFile(entry.Name())
		if readErr != nil {
			t.Fatalf("read %s: %v", entry.Name(), readErr)
		}
		for _, r := range string(content) {
			if unicode.Is(unicode.Han, r) {
				t.Fatalf("%s contains Han text; command diagnostics must use English source text", entry.Name())
			}
		}
	}
}

// TestHostSQLDirsFollowConvention verifies the init and mock SQL helpers keep
// using the expected manifest directory layout.
func TestHostSQLDirsFollowConvention(t *testing.T) {
	t.Parallel()

	if got := hostInitSQLDir(); got != "manifest/sql" {
		t.Fatalf("expected init sql dir %q, got %q", "manifest/sql", got)
	}
	if got := hostMockSQLDir(); got != path.Join("manifest/sql", "mock-data") {
		t.Fatalf("expected mock sql dir %q, got %q", path.Join("manifest/sql", "mock-data"), got)
	}
}

// TestResolveSQLAssetSource verifies the command source selection is explicit
// and defaults to embedded assets for runtime execution.
func TestResolveSQLAssetSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    sqlAssetSource
		wantErr bool
	}{
		{name: "default embedded", input: "", want: sqlAssetSourceEmbedded},
		{name: "explicit embedded", input: "embedded", want: sqlAssetSourceEmbedded},
		{name: "explicit local", input: "local", want: sqlAssetSourceLocal},
		{name: "reject unknown", input: "filesystem", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveSQLAssetSource(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve source: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

// TestParseInitRebuildFlag verifies the optional rebuild flag accepts common
// boolean spellings and rejects ambiguous values.
func TestParseInitRebuildFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{name: "empty defaults to false", input: "", want: false},
		{name: "true enables rebuild", input: "true", want: true},
		{name: "one enables rebuild", input: "1", want: true},
		{name: "yes enables rebuild", input: "yes", want: true},
		{name: "false disables rebuild", input: "false", want: false},
		{name: "zero disables rebuild", input: "0", want: false},
		{name: "no disables rebuild", input: "no", want: false},
		{name: "reject unknown value", input: "maybe", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseInitRebuildFlag(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parse rebuild flag: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestLogHTTPStartupSummaryEmitsFieldsWithoutSQL verifies startup observability
// uses an aggregate summary instead of ORM SQL text.
func TestLogHTTPStartupSummaryEmitsFieldsWithoutSQL(t *testing.T) {
	ctx := context.Background()
	collector := startupstats.New()
	collector.Add(startupstats.CounterCatalogSnapshotBuilds, 1)
	collector.Add(startupstats.CounterIntegrationSnapshotBuilds, 1)
	collector.Add(startupstats.CounterJobSnapshotBuilds, 1)
	collector.Add(startupstats.CounterPluginScans, 1)
	collector.Add(startupstats.CounterPluginSyncChanged, 2)
	collector.Add(startupstats.CounterPluginSyncNoop, 3)
	collector.RecordPhase(startupstats.PhasePluginBootstrapAutoEnable, 12)
	collector.RecordPhase(startupstats.PhasePluginStartupConsistency, 4)

	var logs []string
	logger.Logger().SetHandlers(func(ctx context.Context, in *glog.HandlerInput) {
		logs = append(logs, in.ValuesContent())
	})
	t.Cleanup(func() {
		logger.Logger().SetHandlers()
	})

	logHTTPStartupSummary(ctx, collector)

	joined := strings.Join(logs, "\n")
	for _, expected := range []string{
		"startup summary",
		"catalogSnapshots=1",
		"integrationSnapshots=1",
		"jobSnapshots=1",
		"pluginScans=1",
		"pluginChanged=2",
		"pluginNoop=3",
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected startup summary to contain %q, got %q", expected, joined)
		}
	}
	for _, forbidden := range []string{"SHOW FULL COLUMNS", "SELECT ", "INSERT INTO", "UPDATE ", "DELETE "} {
		if strings.Contains(strings.ToUpper(joined), forbidden) {
			t.Fatalf("expected startup summary to omit SQL text %q, got %q", forbidden, joined)
		}
	}
}

// TestValidateHTTPStartupPluginConsistencyPanicsOnInvalidState verifies
// startup consistency failures stop HTTP startup before later phases run.
func TestValidateHTTPStartupPluginConsistencyPanicsOnInvalidState(t *testing.T) {
	ctx := startupstats.WithCollector(context.Background(), startupstats.New())
	pluginSvc := &startupConsistencyFailingPluginService{err: gerror.New("invalid startup state")}

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected startup consistency failure to panic")
		}
		if !pluginSvc.called {
			t.Fatal("expected startup consistency validator to be called")
		}
		snapshot := startupstats.FromContext(ctx).Snapshot()
		if _, ok := snapshot.Phases[startupstats.PhasePluginStartupConsistency]; !ok {
			t.Fatalf("expected startup consistency phase to be recorded, got %#v", snapshot.Phases)
		}
	}()

	if err := validateHTTPStartupPluginConsistency(ctx, pluginSvc); err != nil {
		t.Fatalf("expected panic path before returning error, got %v", err)
	}
}

// startupConsistencyFailingPluginService is a narrow fake for startup runtime tests.
type startupConsistencyFailingPluginService struct {
	called bool
	err    error
}

// ValidateStartupConsistency records the startup validation call and returns the configured error.
func (s *startupConsistencyFailingPluginService) ValidateStartupConsistency(context.Context) error {
	s.called = true
	return s.err
}

// TestExecuteSQLAssetsWithExecutorStopsAfterFirstError verifies execution halts
// at the first failing SQL asset and returns the failing file name.
func TestExecuteSQLAssetsWithExecutorStopsAfterFirstError(t *testing.T) {
	t.Parallel()

	assets := []sqlAsset{
		{Path: "manifest/sql/001-first.sql", Content: "FIRST"},
		{Path: "manifest/sql/002-second.sql", Content: "SECOND"},
		{Path: "manifest/sql/003-third.sql", Content: "THIRD"},
	}

	var executedSQL []string
	err := executeSQLAssetsWithExecutor(context.Background(), assets, func(ctx context.Context, sql string) error {
		executedSQL = append(executedSQL, sql)
		if sql == "SECOND" {
			return errors.New("boom")
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected execution error")
	}
	if !strings.Contains(err.Error(), "002-second.sql") {
		t.Fatalf("expected error %q to contain failing file name", err.Error())
	}
	if !reflect.DeepEqual(executedSQL, []string{"FIRST", "SECOND"}) {
		t.Fatalf("expected executed sql %v, got %v", []string{"FIRST", "SECOND"}, executedSQL)
	}
}

// TestExecuteSQLAssetsWithExecutorSkipsEmptyFiles verifies blank SQL assets are
// ignored while non-empty assets still execute in order.
func TestExecuteSQLAssetsWithExecutorSkipsEmptyFiles(t *testing.T) {
	t.Parallel()

	assets := []sqlAsset{
		{Path: "manifest/sql/001-empty.sql", Content: ""},
		{Path: "manifest/sql/002-seed.sql", Content: "SEED"},
	}

	var executedSQL []string
	err := executeSQLAssetsWithExecutor(context.Background(), assets, func(ctx context.Context, sql string) error {
		executedSQL = append(executedSQL, sql)
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !reflect.DeepEqual(executedSQL, []string{"SEED"}) {
		t.Fatalf("expected executed sql %v, got %v", []string{"SEED"}, executedSQL)
	}
}

// TestMockCommandFailsWithoutInitializedSQLiteSchema verifies mock-data loading
// depends on an initialized database schema instead of creating tables itself.
func TestMockCommandFailsWithoutInitializedSQLiteSchema(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "linapro.db")
	writeTestSQLFile(
		t,
		filepath.Join(tempDir, "manifest", "sql", "mock-data", "001-users.sql"),
		"INSERT INTO sys_user(username) VALUES ('demo') ON CONFLICT DO NOTHING;",
	)

	adapter, err := gcfg.NewAdapterContent(`
database:
  default:
    link: "sqlite::@file(` + dbPath + `)"
`)
	if err != nil {
		t.Fatalf("create config adapter: %v", err)
	}
	db, err := gdb.New(gdb.ConfigNode{Link: "sqlite::@file(" + dbPath + ")"})
	if err != nil {
		t.Fatalf("open SQLite command DB: %v", err)
	}
	originalAdapter := g.Cfg().GetAdapter()
	originalCommandDatabase := commandDatabase
	g.Cfg().SetAdapter(adapter)
	commandDatabase = func() gdb.DB {
		return db
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current working directory: %v", err)
	}
	if err = os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(cwd); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
		commandDatabase = originalCommandDatabase
		g.Cfg().SetAdapter(originalAdapter)
		if closeErr := db.Close(ctx); closeErr != nil {
			t.Fatalf("close SQLite command DB: %v", closeErr)
		}
	})

	_, err = (&Main{}).Mock(ctx, MockInput{
		Confirm:   mockCommandName,
		SQLSource: string(sqlAssetSourceLocal),
	})
	if err == nil {
		t.Fatal("expected mock SQL to fail when sys_user has not been initialized")
	}
	if !strings.Contains(err.Error(), "001-users.sql") {
		t.Fatalf("expected error to contain mock SQL file name, got %q", err.Error())
	}

	count, err := db.GetValue(ctx, "SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name='sys_user'")
	if err != nil {
		t.Fatalf("inspect SQLite schema after failed mock SQL: %v", err)
	}
	if count.Int() != 0 {
		t.Fatalf("expected mock SQL not to create sys_user table, got count=%d", count.Int())
	}
}

// TestSQLiteInitThenMockCommandLoadsEmbeddedAssets verifies the runtime init
// and mock commands can execute the full embedded host SQL asset set on SQLite.
func TestSQLiteInitThenMockCommandLoadsEmbeddedAssets(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "linapro.db")

	adapter, err := gcfg.NewAdapterContent(`
database:
  default:
    link: "sqlite::@file(` + dbPath + `)"
`)
	if err != nil {
		t.Fatalf("create config adapter: %v", err)
	}
	db, err := gdb.New(gdb.ConfigNode{Link: "sqlite::@file(" + dbPath + ")"})
	if err != nil {
		t.Fatalf("open SQLite command DB: %v", err)
	}
	originalAdapter := g.Cfg().GetAdapter()
	originalCommandDatabase := commandDatabase
	g.Cfg().SetAdapter(adapter)
	commandDatabase = func() gdb.DB {
		return db
	}
	t.Cleanup(func() {
		commandDatabase = originalCommandDatabase
		g.Cfg().SetAdapter(originalAdapter)
		if closeErr := db.Close(ctx); closeErr != nil {
			t.Fatalf("close SQLite command DB: %v", closeErr)
		}
	})

	_, err = (&Main{}).Init(ctx, InitInput{
		Confirm:   initCommandName,
		SQLSource: string(sqlAssetSourceEmbedded),
		Rebuild:   "true",
	})
	if err != nil {
		t.Fatalf("run SQLite init with embedded SQL assets: %v", err)
	}
	_, err = (&Main{}).Mock(ctx, MockInput{
		Confirm:   mockCommandName,
		SQLSource: string(sqlAssetSourceEmbedded),
	})
	if err != nil {
		t.Fatalf("run SQLite mock with embedded SQL assets: %v", err)
	}

	assertSQLiteTableExists(t, ctx, db, "sys_user")
	assertSQLiteUserExists(t, ctx, db, "admin")
	assertSQLiteUserExists(t, ctx, db, "user001")
}

// TestScanLocalSQLAssetsSortsFiles verifies development-mode local SQL loading
// keeps lexical order.
func TestScanLocalSQLAssetsSortsFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sqlDir := filepath.Join(tempDir, "manifest", "sql")
	writeTestSQLFile(t, filepath.Join(sqlDir, "010-third.sql"), "THIRD")
	writeTestSQLFile(t, filepath.Join(sqlDir, "001-first.sql"), "FIRST")
	writeTestSQLFile(t, filepath.Join(sqlDir, "002-second.sql"), "SECOND")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err = os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(cwd); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	}()

	assets, err := scanLocalSQLAssets(context.Background(), hostInitSQLDir())
	if err != nil {
		t.Fatalf("scan local sql assets: %v", err)
	}
	got := []string{assets[0].Content, assets[1].Content, assets[2].Content}
	if !reflect.DeepEqual(got, []string{"FIRST", "SECOND", "THIRD"}) {
		t.Fatalf("expected ordered contents %v, got %v", []string{"FIRST", "SECOND", "THIRD"}, got)
	}
}

// TestScanEmbeddedSQLAssetsReadsPreparedFiles verifies runtime-mode SQL loading
// reads packaged manifest assets from the embedded filesystem.
func TestScanEmbeddedSQLAssetsReadsPreparedFiles(t *testing.T) {
	t.Parallel()

	assets, err := scanEmbeddedSQLAssets(context.Background(), hostInitSQLDir())
	if err != nil {
		t.Fatalf("scan embedded sql assets: %v", err)
	}
	if len(assets) == 0 {
		t.Fatal("expected embedded init sql assets")
	}
	if assets[0].Path != path.Join("manifest/sql", "001-user-auth-bootstrap.sql") {
		t.Fatalf("expected first embedded sql asset %q, got %q", path.Join("manifest/sql", "001-user-auth-bootstrap.sql"), assets[0].Path)
	}
}

// TestInitRuntimeDefaultUsesEmbeddedAssets verifies runtime `lina init`
// behavior defaults to the embedded manifest SQL assets.
func TestInitRuntimeDefaultUsesEmbeddedAssets(t *testing.T) {
	t.Parallel()

	source, err := resolveSQLAssetSource("")
	if err != nil {
		t.Fatalf("resolve default init source: %v", err)
	}

	assets, err := scanInitSQLAssets(context.Background(), source)
	if err != nil {
		t.Fatalf("scan init sql assets: %v", err)
	}
	if len(assets) == 0 {
		t.Fatal("expected embedded init sql assets")
	}
	if assets[0].Path != path.Join("manifest/sql", "001-user-auth-bootstrap.sql") {
		t.Fatalf("expected first embedded init sql asset %q, got %q", path.Join("manifest/sql", "001-user-auth-bootstrap.sql"), assets[0].Path)
	}
}

// TestMockRuntimeDefaultUsesEmbeddedAssets verifies runtime `lina mock`
// behavior defaults to the embedded mock-data SQL assets.
func TestMockRuntimeDefaultUsesEmbeddedAssets(t *testing.T) {
	t.Parallel()

	source, err := resolveSQLAssetSource("")
	if err != nil {
		t.Fatalf("resolve default mock source: %v", err)
	}

	assets, err := scanMockSQLAssets(context.Background(), source)
	if err != nil {
		t.Fatalf("scan mock sql assets: %v", err)
	}
	if len(assets) == 0 {
		t.Fatal("expected embedded mock sql assets")
	}
	if assets[0].Path != path.Join("manifest/sql", "mock-data", "001-users.sql") {
		t.Fatalf(
			"expected first embedded mock sql asset %q, got %q",
			path.Join("manifest/sql", "mock-data", "001-users.sql"),
			assets[0].Path,
		)
	}
}

// assertSQLiteTableExists verifies one table was created in the SQLite schema.
func assertSQLiteTableExists(t *testing.T, ctx context.Context, db gdb.DB, tableName string) {
	t.Helper()

	count, err := db.GetValue(ctx, "SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name=?", tableName)
	if err != nil {
		t.Fatalf("inspect SQLite table %s: %v", tableName, err)
	}
	if count.Int() != 1 {
		t.Fatalf("expected SQLite table %s to exist, got count=%d", tableName, count.Int())
	}
}

// assertSQLiteUserExists verifies one expected seed or mock user was loaded.
func assertSQLiteUserExists(t *testing.T, ctx context.Context, db gdb.DB, username string) {
	t.Helper()

	count, err := db.GetValue(ctx, "SELECT COUNT(1) FROM sys_user WHERE username=?", username)
	if err != nil {
		t.Fatalf("query SQLite user %s: %v", username, err)
	}
	if count.Int() != 1 {
		t.Fatalf("expected SQLite user %s to exist, got count=%d", username, count.Int())
	}
}

// writeTestSQLFile writes one temporary SQL file for shared command helper tests.
func writeTestSQLFile(t *testing.T, path string, contents string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}
