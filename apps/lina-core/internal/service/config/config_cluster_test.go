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
  coordination: redis
  election:
    lease: 45s
    renewInterval: 15s
  redis:
    address: "127.0.0.1:6379"
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
	if cfg.Coordination != ClusterCoordinationRedis {
		t.Fatalf("expected redis coordination, got %q", cfg.Coordination)
	}
	if cfg.Redis.Address != "127.0.0.1:6379" {
		t.Fatalf("expected redis address to be loaded, got %q", cfg.Redis.Address)
	}
	if cfg.Redis.ConnectTimeout != 3*time.Second ||
		cfg.Redis.ReadTimeout != 2*time.Second ||
		cfg.Redis.WriteTimeout != 2*time.Second {
		t.Fatalf("expected default redis timeouts, got connect=%s read=%s write=%s", cfg.Redis.ConnectTimeout, cfg.Redis.ReadTimeout, cfg.Redis.WriteTimeout)
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
  coordination: redis
  redis:
    address: "127.0.0.1:6379"
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
  coordination: redis
  redis:
    address: "127.0.0.1:6379"
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

// TestPostgreSQLDialectOnStartupKeepsConfigServiceClusterEnabled verifies
// PostgreSQL startup hooks are no-op for cluster mode and SQLite warnings.
func TestPostgreSQLDialectOnStartupKeepsConfigServiceClusterEnabled(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  coordination: redis
  redis:
    address: "127.0.0.1:6379"
`)

	svc := New()
	var warnings []string
	logger.Logger().SetHandlers(func(ctx context.Context, in *glog.HandlerInput) {
		warnings = append(warnings, in.ValuesContent())
	})
	t.Cleanup(func() {
		logger.Logger().SetHandlers()
	})

	dbDialect, err := dialect.From("pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable")
	if err != nil {
		t.Fatalf("resolve PostgreSQL dialect failed: %v", err)
	}
	if err = dbDialect.OnStartup(context.Background(), svc); err != nil {
		t.Fatalf("run PostgreSQL startup hook failed: %v", err)
	}

	if !svc.IsClusterEnabled(context.Background()) {
		t.Fatal("expected PostgreSQL startup hook to preserve enabled cluster mode")
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no PostgreSQL startup warnings, got %#v", warnings)
	}
}

// TestGetClusterPanicsWhenCoordinationMissing verifies clustered deployment
// requires an explicit coordination backend.
func TestGetClusterPanicsWhenCoordinationMissing(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  redis:
    address: "127.0.0.1:6379"
`)

	defer assertConfigPanicContains(t, "field=cluster.coordination")
	New().GetCluster(context.Background())
}

// TestGetClusterPanicsWhenCoordinationUnsupported verifies only Redis is
// accepted as the current coordination backend.
func TestGetClusterPanicsWhenCoordinationUnsupported(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  coordination: postgres
  redis:
    address: "127.0.0.1:6379"
`)

	defer assertConfigPanicContains(t, "fix=set cluster.coordination=redis")
	New().GetCluster(context.Background())
}

// TestGetClusterPanicsWhenRedisAddressMissing verifies Redis address is
// required whenever Redis coordination is enabled.
func TestGetClusterPanicsWhenRedisAddressMissing(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  coordination: redis
`)

	defer assertConfigPanicContains(t, "field=cluster.redis.address")
	New().GetCluster(context.Background())
}

// TestGetClusterPanicsWhenRedisTimeoutInvalid verifies Redis timeout fields
// must be duration strings with units.
func TestGetClusterPanicsWhenRedisTimeoutInvalid(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  coordination: redis
  redis:
    address: "127.0.0.1:6379"
    connectTimeout: invalid
`)

	defer assertConfigPanicContains(t, "parse config cluster.redis.connectTimeout failed")
	New().GetCluster(context.Background())
}

// TestSingleNodeModeDoesNotRequireRedis verifies local deployments can omit all
// Redis coordination settings.
func TestSingleNodeModeDoesNotRequireRedis(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: false
`)

	cfg := New().GetCluster(context.Background())
	if cfg.Enabled {
		t.Fatal("expected cluster mode disabled")
	}
	if cfg.Redis.Address != "" {
		t.Fatalf("expected empty redis address in single-node config, got %q", cfg.Redis.Address)
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
