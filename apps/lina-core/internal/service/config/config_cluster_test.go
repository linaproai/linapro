package config

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

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

func setTestConfigContent(t *testing.T, content string) {
	t.Helper()

	adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile)
	if !ok {
		t.Fatal("expected config adapter to be *gcfg.AdapterFile")
	}

	originalContent := adapter.GetContent()
	adapter.SetContent(content)

	t.Cleanup(func() {
		if originalContent != "" {
			adapter.SetContent(originalContent)
			return
		}
		adapter.RemoveContent()
	})
}
