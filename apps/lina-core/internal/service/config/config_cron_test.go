// This file verifies scheduled-job runtime configuration parsing and platform
// guards for protected sys_config values.

package config

import (
	"context"
	"runtime"
	"testing"
)

// TestGetCronUsesProtectedRuntimeValues verifies cron protected settings flow
// into both structured config and convenience getters.
func TestGetCronUsesProtectedRuntimeValues(t *testing.T) {
	withRuntimeParamValue(t, RuntimeParamKeyCronShellEnabled, "true")
	withRuntimeParamValue(t, RuntimeParamKeyCronLogRetention, `{"mode":"count","value":500}`)

	svc := New()
	cfg := svc.GetCron(context.Background())

	if cfg.LogRetention.Mode != CronLogRetentionModeCount {
		t.Fatalf("expected cron log retention mode count, got %q", cfg.LogRetention.Mode)
	}
	if cfg.LogRetention.Value != 500 {
		t.Fatalf("expected cron log retention value 500, got %d", cfg.LogRetention.Value)
	}
	if retention := svc.GetCronLogRetention(context.Background()); retention.Mode != CronLogRetentionModeCount || retention.Value != 500 {
		t.Fatalf("expected cron log retention getter to return count/500, got mode=%q value=%d", retention.Mode, retention.Value)
	}

	expectedShellEnabled := buildCronShellConfig(true, runtime.GOOS).Enabled
	if cfg.Shell.Enabled != expectedShellEnabled {
		t.Fatalf("expected cron shell enabled to be %t, got %t", expectedShellEnabled, cfg.Shell.Enabled)
	}
	if shellEnabled := svc.IsCronShellEnabled(context.Background()); shellEnabled != expectedShellEnabled {
		t.Fatalf("expected IsCronShellEnabled to be %t, got %t", expectedShellEnabled, shellEnabled)
	}
}

// TestGetCronUsesDefaultRetentionWhenRuntimeParamMissing verifies the host
// falls back to the built-in cleanup policy when sys_config rows are absent.
func TestGetCronUsesDefaultRetentionWhenRuntimeParamMissing(t *testing.T) {
	withRuntimeParamAbsent(t, RuntimeParamKeyCronShellEnabled)
	withRuntimeParamAbsent(t, RuntimeParamKeyCronLogRetention)

	cfg := New().GetCron(context.Background())
	if cfg.LogRetention.Mode != CronLogRetentionModeDays {
		t.Fatalf("expected default cron log retention mode days, got %q", cfg.LogRetention.Mode)
	}
	if cfg.LogRetention.Value != 30 {
		t.Fatalf("expected default cron log retention value 30, got %d", cfg.LogRetention.Value)
	}
	if cfg.Shell.Enabled {
		t.Fatal("expected shell mode to stay disabled by default")
	}
}

// TestBuildCronShellConfigDisablesShellOnWindows verifies platform guard logic
// forces shell jobs off on Windows regardless of the stored switch value.
func TestBuildCronShellConfigDisablesShellOnWindows(t *testing.T) {
	windowsCfg := buildCronShellConfig(true, "windows")
	if windowsCfg.Enabled {
		t.Fatal("expected windows shell config to be disabled")
	}
	if windowsCfg.Supported {
		t.Fatal("expected windows shell config to report unsupported")
	}
	if windowsCfg.DisabledReason != cronShellUnsupportedReason {
		t.Fatalf("expected windows disabled reason %q, got %q", cronShellUnsupportedReason, windowsCfg.DisabledReason)
	}

	linuxCfg := buildCronShellConfig(true, "linux")
	if !linuxCfg.Enabled {
		t.Fatal("expected linux shell config to stay enabled")
	}
	if !linuxCfg.Supported {
		t.Fatal("expected linux shell config to report supported")
	}
	if linuxCfg.DisabledReason != "" {
		t.Fatalf("expected linux disabled reason empty, got %q", linuxCfg.DisabledReason)
	}
}
