// This file verifies upload-related configuration loading and runtime
// overrides.

package config

import (
	"context"
	"testing"
)

// TestGetUploadUsesDefaultWhenUnset verifies upload config falls back to its
// defaults when static config and runtime overrides are absent.
func TestGetUploadUsesDefaultWhenUnset(t *testing.T) {
	setTestConfigContent(t, `
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
`)
	withRuntimeParamAbsent(t, RuntimeParamKeyUploadMaxSize)

	svc := New()
	cfg, err := svc.GetUpload(context.Background())
	if err != nil {
		t.Fatalf("get upload config: %v", err)
	}

	if cfg.Path != defaultUploadPath {
		t.Fatalf("expected default upload path %q, got %q", defaultUploadPath, cfg.Path)
	}
	if cfg.MaxSize != defaultUploadMaxSize {
		t.Fatalf("expected default upload max size %d, got %d", defaultUploadMaxSize, cfg.MaxSize)
	}
}

// TestGetUploadPathUsesStaticConfig verifies static upload settings remain
// available when runtime overrides are absent.
func TestGetUploadPathUsesStaticConfig(t *testing.T) {
	setTestConfigContent(t, `
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
upload:
  path: runtime/uploads
  maxSize: 32
`)
	withRuntimeParamAbsent(t, RuntimeParamKeyUploadMaxSize)

	svc := New()
	if path := svc.GetUploadPath(context.Background()); path != resolveRuntimePath("runtime/uploads") {
		t.Fatalf("expected upload path to be runtime/uploads, got %s", path)
	}

	cfg, err := svc.GetUpload(context.Background())
	if err != nil {
		t.Fatalf("get upload config: %v", err)
	}
	if cfg.Path != "runtime/uploads" {
		t.Fatalf("expected upload config path to be runtime/uploads, got %s", cfg.Path)
	}
	if cfg.MaxSize != 32 {
		t.Fatalf("expected upload config max size to be 32, got %d", cfg.MaxSize)
	}
	maxSize, err := svc.GetUploadMaxSize(context.Background())
	if err != nil {
		t.Fatalf("get upload max size: %v", err)
	}
	if maxSize != 32 {
		t.Fatalf("expected upload runtime getter max size to be 32, got %d", maxSize)
	}
}

// TestGetUploadPrefersRuntimeParamMaxSize verifies runtime upload size
// overrides flow into both structured config and convenience getters.
func TestGetUploadPrefersRuntimeParamMaxSize(t *testing.T) {
	withRuntimeParamValue(t, RuntimeParamKeyUploadMaxSize, "8")

	svc := New()
	cfg, err := svc.GetUpload(context.Background())
	if err != nil {
		t.Fatalf("get upload config: %v", err)
	}

	if cfg.MaxSize != 8 {
		t.Fatalf("expected runtime param upload max size to be 8, got %d", cfg.MaxSize)
	}
	maxSize, err := svc.GetUploadMaxSize(context.Background())
	if err != nil {
		t.Fatalf("get upload max size: %v", err)
	}
	if maxSize != 8 {
		t.Fatalf("expected runtime getter upload max size to be 8, got %d", maxSize)
	}
}
