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

// TestInitDatabaseLinkHelpers verifies init can derive the target schema and a
// server-level connection link from the configured GoFrame MySQL link.
func TestInitDatabaseLinkHelpers(t *testing.T) {
	t.Parallel()

	link := "mysql:root:12345678@tcp(127.0.0.1:3306)/linapro?charset=utf8mb4&parseTime=true&loc=Local"
	name, err := databaseNameFromMySQLLink(link)
	if err != nil {
		t.Fatalf("extract database name: %v", err)
	}
	if name != initDatabaseName {
		t.Fatalf("expected database name %q, got %q", initDatabaseName, name)
	}

	serverLink, err := serverLinkFromMySQLLink(link)
	if err != nil {
		t.Fatalf("derive server link: %v", err)
	}
	wantServerLink := "mysql:root:12345678@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=true&loc=Local"
	if serverLink != wantServerLink {
		t.Fatalf("expected server link %q, got %q", wantServerLink, serverLink)
	}
}

// TestInitDatabaseLinkHelpersRejectMissingDatabase verifies init refuses links
// that cannot identify the target schema to create or rebuild.
func TestInitDatabaseLinkHelpersRejectMissingDatabase(t *testing.T) {
	t.Parallel()

	for _, link := range []string{
		"mysql:root:12345678@tcp(127.0.0.1:3306)",
		"mysql:root:12345678@tcp(127.0.0.1:3306)/?charset=utf8mb4",
	} {
		link := link
		t.Run(link, func(t *testing.T) {
			t.Parallel()

			if _, err := databaseNameFromMySQLLink(link); err == nil {
				t.Fatal("expected database name extraction error")
			}
			if _, err := serverLinkFromMySQLLink(link); err == nil {
				t.Fatal("expected server link extraction error")
			}
		})
	}
}

// TestQuoteMySQLIdentifier verifies bootstrap SQL escapes MySQL identifiers
// instead of concatenating raw names.
func TestQuoteMySQLIdentifier(t *testing.T) {
	t.Parallel()

	got, err := quoteMySQLIdentifier("lina`pro")
	if err != nil {
		t.Fatalf("quote identifier: %v", err)
	}
	if got != "`lina``pro`" {
		t.Fatalf("expected escaped identifier, got %q", got)
	}
	if _, err = quoteMySQLIdentifier(""); err == nil {
		t.Fatal("expected empty identifier error")
	}
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
	if assets[0].Path != path.Join("manifest/sql", "001-project-init.sql") {
		t.Fatalf("expected first embedded sql asset %q, got %q", path.Join("manifest/sql", "001-project-init.sql"), assets[0].Path)
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
	if assets[0].Path != path.Join("manifest/sql", "001-project-init.sql") {
		t.Fatalf("expected first embedded init sql asset %q, got %q", path.Join("manifest/sql", "001-project-init.sql"), assets[0].Path)
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
	if assets[0].Path != path.Join("manifest/sql", "mock-data", "003-mock-users.sql") {
		t.Fatalf(
			"expected first embedded mock sql asset %q, got %q",
			path.Join("manifest/sql", "mock-data", "003-mock-users.sql"),
			assets[0].Path,
		)
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
