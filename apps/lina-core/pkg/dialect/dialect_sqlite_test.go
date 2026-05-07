// This file tests SQLite database preparation, link parsing, and startup hooks.

package dialect

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/os/glog"

	"lina-core/pkg/logger"
)

// fakeRuntimeConfig records the cluster override requested by startup hooks.
type fakeRuntimeConfig struct {
	called bool
	value  bool
}

// OverrideClusterEnabledForDialect records the requested cluster override.
func (f *fakeRuntimeConfig) OverrideClusterEnabledForDialect(value bool) {
	f.called = true
	f.value = value
}

// TestSQLitePathFromLink verifies GoFrame SQLite file links are parsed
// without accepting unsupported shorthand forms.
// TestSQLitePrepareDatabaseRejectsTildePath verifies unsupported shell-style
// home expansion is rejected with a clear error through the public contract.
func TestSQLitePrepareDatabaseRejectsTildePath(t *testing.T) {
	t.Parallel()

	dbDialect, err := From("sqlite::@file(~/.linapro/data/linapro.db)")
	if err != nil {
		t.Fatalf("resolve SQLite dialect failed: %v", err)
	}
	err = dbDialect.PrepareDatabase(context.Background(), "sqlite::@file(~/.linapro/data/linapro.db)", false)
	if err == nil {
		t.Fatal("expected tilde SQLite path to fail")
	}
}

// TestSQLiteOnStartupOverridesCluster verifies SQLite startup hooks force the
// runtime cluster flag off through the stable narrow interface.
func TestSQLiteOnStartupOverridesCluster(t *testing.T) {
	runtime := &fakeRuntimeConfig{}
	dbDialect, err := From("sqlite::@file(./temp/sqlite/linapro.db)")
	if err != nil {
		t.Fatalf("resolve SQLite dialect failed: %v", err)
	}
	var warnings []string
	logger.Logger().SetHandlers(func(ctx context.Context, in *glog.HandlerInput) {
		warnings = append(warnings, in.ValuesContent())
	})
	t.Cleanup(func() {
		logger.Logger().SetHandlers()
	})

	if err = dbDialect.OnStartup(context.Background(), runtime); err != nil {
		t.Fatalf("run SQLite startup hook failed: %v", err)
	}
	if !runtime.called {
		t.Fatal("expected SQLite startup hook to override cluster mode")
	}
	if runtime.value {
		t.Fatal("expected SQLite startup hook to force cluster mode off")
	}
	if len(warnings) != 4 {
		t.Fatalf("expected 4 SQLite warnings, got %d: %#v", len(warnings), warnings)
	}
	for _, needle := range []string{
		"[WARNING]",
		"sqlite::@file(./temp/sqlite/linapro.db)",
		"cluster.enabled",
		"production",
		"MySQL",
	} {
		if !containsAnyWarning(warnings, needle) {
			t.Fatalf("expected SQLite warning to contain %q, got %#v", needle, warnings)
		}
	}
}

// containsAnyWarning reports whether one captured warning contains a substring.
func containsAnyWarning(warnings []string, needle string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, needle) {
			return true
		}
	}
	return false
}
