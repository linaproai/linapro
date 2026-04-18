// This file verifies logger configuration loading for structured logging.

package config

import (
	"context"
	"testing"
)

// TestGetLoggerUsesStructuredSwitch verifies the custom logger extension switch
// is read from config content.
func TestGetLoggerUsesStructuredSwitch(t *testing.T) {
	setTestConfigContent(t, `
logger:
  path: "/tmp/lina"
  file: "lina-{Y-m-d}.log"
  stdout: false
  extensions:
    structured: true
`)

	cfg := New().GetLogger(context.Background())

	if cfg.Path != "/tmp/lina" {
		t.Fatalf("expected log path to be loaded, got %q", cfg.Path)
	}
	if cfg.File != "lina-{Y-m-d}.log" {
		t.Fatalf("expected log file pattern to be loaded, got %q", cfg.File)
	}
	if cfg.Stdout {
		t.Fatal("expected stdout switch to be disabled")
	}
	if !cfg.Extensions.Structured {
		t.Fatal("expected structured logging switch to be enabled")
	}
}

// TestGetLoggerUsesDefaultWhenSwitchMissing verifies logger defaults remain in
// effect when the extension switch is not configured.
func TestGetLoggerUsesDefaultWhenSwitchMissing(t *testing.T) {
	setTestConfigContent(t, `
logger:
  level: "all"
`)

	cfg := New().GetLogger(context.Background())

	if cfg.Path != "" {
		t.Fatalf("expected default log path to be empty, got %q", cfg.Path)
	}
	if cfg.File != defaultLoggerFilePattern {
		t.Fatalf("expected default log file pattern %q, got %q", defaultLoggerFilePattern, cfg.File)
	}
	if !cfg.Stdout {
		t.Fatal("expected stdout switch to be enabled by default")
	}
	if cfg.Extensions.Structured {
		t.Fatal("expected structured logging switch to be disabled by default")
	}
}
