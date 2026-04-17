// This file verifies logger configuration loading for structured logging.

package config

import (
	"context"
	"testing"
)

func TestGetLoggerUsesStructuredSwitch(t *testing.T) {
	setTestConfigContent(t, `
logger:
  structured: true
`)

	cfg := New().GetLogger(context.Background())

	if !cfg.Structured {
		t.Fatal("expected structured logging switch to be enabled")
	}
}

func TestGetLoggerUsesDefaultWhenSwitchMissing(t *testing.T) {
	setTestConfigContent(t, `
logger:
  level: "all"
`)

	cfg := New().GetLogger(context.Background())

	if cfg.Structured {
		t.Fatal("expected structured logging switch to be disabled by default")
	}
}
