// This file verifies cluster configuration loading and default election
// fallback behavior.

package config

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/glog"

	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
)

// TestGetClusterUsesClusterElectionConfig verifies nested cluster election
// settings are loaded from config content.
func TestGetClusterUsesClusterElectionConfig(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  election:
    lease: 45s
    renewInterval: 15s
`)

	cfg := New().GetCluster(context.Background())

	if !cfg.Enabled {
		t.Fatal("expected cluster mode to be enabled")
	}
	if cfg.Election.Lease != 45*time.Second {
		t.Fatalf("expected cluster lease to be 45s, got %s", cfg.Election.Lease)
	}
	if cfg.Election.RenewInterval != 15*time.Second {
		t.Fatalf("expected cluster renew interval to be 15s, got %s", cfg.Election.RenewInterval)
	}
}

// TestGetClusterUsesDefaultsWhenElectionConfigMissing verifies election timing
// falls back to defaults when the nested section is absent.
func TestGetClusterUsesDefaultsWhenElectionConfigMissing(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: false
`)

	cfg := New().GetCluster(context.Background())

	if cfg.Enabled {
		t.Fatal("expected cluster mode to be disabled")
	}
	if cfg.Election.Lease != 30*time.Second {
		t.Fatalf("expected default lease to be 30s, got %s", cfg.Election.Lease)
	}
	if cfg.Election.RenewInterval != 10*time.Second {
		t.Fatalf("expected default renew interval to be 10s, got %s", cfg.Election.RenewInterval)
	}
}

// TestGetClusterIgnoresLegacyElectionConfig verifies the deprecated root-level
// election section no longer affects cluster defaults.
func TestGetClusterIgnoresLegacyElectionConfig(t *testing.T) {
	setTestConfigContent(t, `
election:
  lease: 50s
  renewInterval: 20s
`)

	cfg := New().GetCluster(context.Background())

	if cfg.Enabled {
		t.Fatal("expected cluster mode to remain disabled by default")
	}
	if cfg.Election.Lease != 30*time.Second {
		t.Fatalf("expected default lease to remain 30s, got %s", cfg.Election.Lease)
	}
	if cfg.Election.RenewInterval != 10*time.Second {
		t.Fatalf("expected default renew interval to remain 10s, got %s", cfg.Election.RenewInterval)
	}
}

// TestOverrideClusterEnabledForDialect verifies a dialect can lock cluster
// mode off in memory regardless of the configured cluster.enabled value.
func TestOverrideClusterEnabledForDialect(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  election:
    lease: 45s
    renewInterval: 15s
`)

	svc := New()
	if !svc.IsClusterEnabled(context.Background()) {
		t.Fatal("expected config to enable cluster mode before dialect override")
	}

	svc.OverrideClusterEnabledForDialect(false)
	if svc.IsClusterEnabled(context.Background()) {
		t.Fatal("expected dialect override to force cluster mode off")
	}

	cfg := svc.GetCluster(context.Background())
	if cfg.Enabled {
		t.Fatal("expected GetCluster to reflect dialect cluster override")
	}
	if cfg.Election.Lease != 45*time.Second {
		t.Fatalf("expected election lease to be preserved, got %s", cfg.Election.Lease)
	}
}

// TestSQLiteDialectOnStartupOverridesConfigService verifies the concrete config
// service satisfies dialect.RuntimeConfig during startup.
func TestSQLiteDialectOnStartupOverridesConfigService(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  election:
    lease: 45s
    renewInterval: 15s
`)

	svc := New()
	var messages []string
	logger.Logger().SetHandlers(func(ctx context.Context, in *glog.HandlerInput) {
		messages = append(messages, in.ValuesContent())
	})
	t.Cleanup(func() {
		logger.Logger().SetHandlers()
	})

	dbDialect, err := dialect.From("sqlite::@file(./temp/sqlite/linapro.db)")
	if err != nil {
		t.Fatalf("resolve SQLite dialect failed: %v", err)
	}
	if err = dbDialect.OnStartup(context.Background(), svc); err != nil {
		t.Fatalf("run SQLite startup hook failed: %v", err)
	}

	if svc.IsClusterEnabled(context.Background()) {
		t.Fatal("expected SQLite startup hook to force config service cluster mode off")
	}
	if len(messages) != 3 {
		t.Fatalf("expected 3 SQLite startup messages, got %d: %#v", len(messages), messages)
	}
	for _, needle := range []string{
		"sqlite::@file(./temp/sqlite/linapro.db)",
		"cluster.enabled",
		"production",
	} {
		if !containsCapturedWarning(messages, needle) {
			t.Fatalf("expected SQLite startup message to contain %q, got %#v", needle, messages)
		}
	}
}

// TestMySQLDialectOnStartupKeepsConfigServiceClusterEnabled verifies MySQL
// startup hooks are no-op for cluster mode and SQLite warnings.
func TestMySQLDialectOnStartupKeepsConfigServiceClusterEnabled(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
`)

	svc := New()
	var warnings []string
	logger.Logger().SetHandlers(func(ctx context.Context, in *glog.HandlerInput) {
		warnings = append(warnings, in.ValuesContent())
	})
	t.Cleanup(func() {
		logger.Logger().SetHandlers()
	})

	dbDialect, err := dialect.From("mysql:root:12345678@tcp(127.0.0.1:3306)/linapro")
	if err != nil {
		t.Fatalf("resolve MySQL dialect failed: %v", err)
	}
	if err = dbDialect.OnStartup(context.Background(), svc); err != nil {
		t.Fatalf("run MySQL startup hook failed: %v", err)
	}

	if !svc.IsClusterEnabled(context.Background()) {
		t.Fatal("expected MySQL startup hook to preserve enabled cluster mode")
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no MySQL startup warnings, got %#v", warnings)
	}
}

// containsCapturedWarning reports whether one captured warning contains a substring.
func containsCapturedWarning(warnings []string, needle string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, needle) {
			return true
		}
	}
	return false
}

// setTestConfigContent swaps the config adapter content for one test case and
// restores it afterward.
func setTestConfigContent(t *testing.T, content string) {
	t.Helper()

	adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile)
	if !ok {
		t.Fatal("expected config adapter to be *gcfg.AdapterFile")
	}

	originalContent := adapter.GetContent()
	adapter.SetContent(content)
	resetStaticConfigCaches()

	t.Cleanup(func() {
		if originalContent != "" {
			adapter.SetContent(originalContent)
		} else {
			adapter.RemoveContent()
		}
		resetStaticConfigCaches()
	})
}
