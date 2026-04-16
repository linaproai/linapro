// This file verifies duration-based configuration parsing for JWT, session,
// and monitor settings.

package config

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDurationConfigsUseDefaultsWhenUnset(t *testing.T) {
	setTestConfigContent(t, `
test: 1
`)

	var (
		jwtCfg     = New().GetJwt(context.Background())
		sessionCfg = New().GetSession(context.Background())
		monitorCfg = New().GetMonitor(context.Background())
	)

	if jwtCfg.Expire != 24*time.Hour {
		t.Fatalf("expected default jwt expire to be 24h, got %s", jwtCfg.Expire)
	}
	if sessionCfg.Timeout != 24*time.Hour {
		t.Fatalf("expected default session timeout to be 24h, got %s", sessionCfg.Timeout)
	}
	if sessionCfg.CleanupInterval != 5*time.Minute {
		t.Fatalf("expected default session cleanup interval to be 5m, got %s", sessionCfg.CleanupInterval)
	}
	if monitorCfg.Interval != 30*time.Second {
		t.Fatalf("expected default monitor interval to be 30s, got %s", monitorCfg.Interval)
	}
	if monitorCfg.RetentionMultiplier != 5 {
		t.Fatalf("expected default retention multiplier to be 5, got %d", monitorCfg.RetentionMultiplier)
	}
}

func TestGetJwtUsesDurationConfig(t *testing.T) {
	setTestConfigContent(t, `
jwt:
  secret: "test-secret"
  expire: 36h
`)

	cfg := New().GetJwt(context.Background())

	if cfg.Secret != "test-secret" {
		t.Fatalf("expected jwt secret to be loaded, got %q", cfg.Secret)
	}
	if cfg.Expire != 36*time.Hour {
		t.Fatalf("expected jwt expire to be 36h, got %s", cfg.Expire)
	}
}

func TestGetSessionUsesDurationConfig(t *testing.T) {
	setTestConfigContent(t, `
session:
  timeout: 36h
  cleanupInterval: 10m
`)

	cfg := New().GetSession(context.Background())

	if cfg.Timeout != 36*time.Hour {
		t.Fatalf("expected session timeout to be 36h, got %s", cfg.Timeout)
	}
	if cfg.CleanupInterval != 10*time.Minute {
		t.Fatalf("expected session cleanup interval to be 10m, got %s", cfg.CleanupInterval)
	}
}

func TestGetMonitorUsesDurationConfigAndRetentionMultiplier(t *testing.T) {
	setTestConfigContent(t, `
monitor:
  interval: 45s
  retentionMultiplier: 8
`)

	cfg := New().GetMonitor(context.Background())

	if cfg.Interval != 45*time.Second {
		t.Fatalf("expected monitor interval to be 45s, got %s", cfg.Interval)
	}
	if cfg.RetentionMultiplier != 8 {
		t.Fatalf("expected retention multiplier to be 8, got %d", cfg.RetentionMultiplier)
	}
}

func TestGetSessionRejectsNonSecondAlignedCleanupInterval(t *testing.T) {
	setTestConfigContent(t, `
session:
  cleanupInterval: 1500ms
`)

	defer assertConfigPanicContains(t, "整秒时长")

	_ = New().GetSession(context.Background())
}

func TestGetMonitorRejectsSubSecondInterval(t *testing.T) {
	setTestConfigContent(t, `
monitor:
  interval: 500ms
`)

	defer assertConfigPanicContains(t, "至少为 1s")

	_ = New().GetMonitor(context.Background())
}

func assertConfigPanicContains(t *testing.T, expected string) {
	t.Helper()

	recovered := recover()
	if recovered == nil {
		t.Fatalf("expected panic containing %q, but no panic occurred", expected)
	}
	if !strings.Contains(recovered.(error).Error(), expected) {
		t.Fatalf("expected panic containing %q, got %v", expected, recovered)
	}
}
