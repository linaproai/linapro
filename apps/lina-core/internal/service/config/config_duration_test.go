// This file verifies duration-based configuration parsing for JWT, session,
// and monitor settings.

package config

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestDurationConfigsUseDefaultsWhenUnset verifies duration-based config getters
// fall back to their baked-in defaults when config is absent.
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

// TestGetJwtUsesDurationConfig verifies JWT duration settings come from static config.
func TestGetJwtUsesDurationConfig(t *testing.T) {
	setTestConfigContent(t, `
database:
  default:
    link: "mysql:root:12345678@tcp(127.0.0.1:3306)/lina?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true"
jwt:
  secret: "test-secret"
  expire: 36h
`)
	withRuntimeParamAbsent(t, RuntimeParamKeyJWTExpire)

	svc := New()
	cfg := svc.GetJwt(context.Background())

	if cfg.Expire != 36*time.Hour {
		t.Fatalf("expected jwt expire to be 36h, got %s", cfg.Expire)
	}
	if cfg.Secret != "test-secret" {
		t.Fatalf("expected jwt secret to be test-secret, got %q", cfg.Secret)
	}
	if expire := svc.GetJwtExpire(context.Background()); expire != 36*time.Hour {
		t.Fatalf("expected GetJwtExpire to be 36h, got %s", expire)
	}
	if secret := svc.GetJwtSecret(context.Background()); secret != "test-secret" {
		t.Fatalf("expected GetJwtSecret to be test-secret, got %q", secret)
	}
}

// TestGetSessionUsesDurationConfig verifies session duration settings come from static config.
func TestGetSessionUsesDurationConfig(t *testing.T) {
	setTestConfigContent(t, `
database:
  default:
    link: "mysql:root:12345678@tcp(127.0.0.1:3306)/lina?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true"
session:
  timeout: 36h
  cleanupInterval: 10m
`)
	withRuntimeParamAbsent(t, RuntimeParamKeySessionTimeout)

	svc := New()
	cfg := svc.GetSession(context.Background())

	if cfg.Timeout != 36*time.Hour {
		t.Fatalf("expected session timeout to be 36h, got %s", cfg.Timeout)
	}
	if cfg.CleanupInterval != 10*time.Minute {
		t.Fatalf("expected session cleanup interval to be 10m, got %s", cfg.CleanupInterval)
	}
	if timeout := svc.GetSessionTimeout(context.Background()); timeout != 36*time.Hour {
		t.Fatalf("expected GetSessionTimeout to be 36h, got %s", timeout)
	}
}

// TestGetMonitorUsesDurationConfigAndRetentionMultiplier verifies monitor
// interval parsing and retention multiplier loading.
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

// TestGetUploadPathUsesStaticConfig verifies static upload settings remain
// available when runtime overrides are absent.
func TestGetUploadPathUsesStaticConfig(t *testing.T) {
	setTestConfigContent(t, `
upload:
  path: runtime/uploads
  maxSize: 32
`)
	withRuntimeParamAbsent(t, RuntimeParamKeyUploadMaxSize)

	svc := New()
	if path := svc.GetUploadPath(context.Background()); path != "runtime/uploads" {
		t.Fatalf("expected upload path to be runtime/uploads, got %s", path)
	}

	cfg := svc.GetUpload(context.Background())
	if cfg.Path != "runtime/uploads" {
		t.Fatalf("expected upload config path to be runtime/uploads, got %s", cfg.Path)
	}
	if cfg.MaxSize != 32 {
		t.Fatalf("expected upload config max size to be 32, got %d", cfg.MaxSize)
	}
	if maxSize := svc.GetUploadMaxSize(context.Background()); maxSize != 32 {
		t.Fatalf("expected upload runtime getter max size to be 32, got %d", maxSize)
	}
}

// TestGetSessionRejectsNonSecondAlignedCleanupInterval verifies invalid
// fractional-second cleanup intervals panic during config load.
func TestGetSessionRejectsNonSecondAlignedCleanupInterval(t *testing.T) {
	setTestConfigContent(t, `
session:
  cleanupInterval: 1500ms
`)

	defer assertConfigPanicContains(t, "整秒时长")

	_ = New().GetSession(context.Background())
}

// TestGetMonitorRejectsSubSecondInterval verifies monitor intervals shorter
// than one second are rejected.
func TestGetMonitorRejectsSubSecondInterval(t *testing.T) {
	setTestConfigContent(t, `
monitor:
  interval: 500ms
`)

	defer assertConfigPanicContains(t, "至少为 1s")

	_ = New().GetMonitor(context.Background())
}

// assertConfigPanicContains verifies the current deferred panic contains the expected text.
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
