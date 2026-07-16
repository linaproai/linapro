// This file verifies upload-related configuration loading and runtime
// overrides.

package config

import (
	"context"
	"testing"
	"time"
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

// TestGetUploadMultipartDefaults verifies multipart planning defaults.
func TestGetUploadMultipartDefaults(t *testing.T) {
	t.Parallel()
	withRuntimeParamAbsent(t, RuntimeParamKeyUploadMultipartEnabled)
	withRuntimeParamAbsent(t, RuntimeParamKeyUploadMultipartThresholdMB)
	withRuntimeParamAbsent(t, RuntimeParamKeyUploadMultipartPartSizeMB)
	withRuntimeParamAbsent(t, RuntimeParamKeyUploadMultipartMaxConcurrency)

	svc := New()
	enabled, err := svc.GetUploadMultipartEnabled(context.Background())
	if err != nil {
		t.Fatalf("GetUploadMultipartEnabled: %v", err)
	}
	if !enabled {
		t.Fatal("expected multipart enabled by default")
	}
	threshold, err := svc.GetUploadMultipartThresholdMB(context.Background())
	if err != nil {
		t.Fatalf("GetUploadMultipartThresholdMB: %v", err)
	}
	if threshold != defaultUploadMultipartThresholdMB {
		t.Fatalf("threshold: got %d want %d", threshold, defaultUploadMultipartThresholdMB)
	}
	partSize, err := svc.GetUploadMultipartPartSizeMB(context.Background())
	if err != nil {
		t.Fatalf("GetUploadMultipartPartSizeMB: %v", err)
	}
	if partSize != defaultUploadMultipartPartSizeMB {
		t.Fatalf("partSize: got %d want %d", partSize, defaultUploadMultipartPartSizeMB)
	}
	concurrency, err := svc.GetUploadMultipartMaxConcurrency(context.Background())
	if err != nil {
		t.Fatalf("GetUploadMultipartMaxConcurrency: %v", err)
	}
	if concurrency != defaultUploadMultipartMaxConcurrency {
		t.Fatalf("concurrency: got %d want %d", concurrency, defaultUploadMultipartMaxConcurrency)
	}
}

// TestValidateUploadMultipartPartSizeConfigValue rejects values below 5MB.
func TestValidateUploadMultipartPartSizeConfigValue(t *testing.T) {
	t.Parallel()
	if err := validateUploadMultipartPartSizeConfigValue(RuntimeParamKeyUploadMultipartPartSizeMB, "4"); err == nil {
		t.Fatal("expected part size 4 to be rejected")
	}
	if err := validateUploadMultipartPartSizeConfigValue(RuntimeParamKeyUploadMultipartPartSizeMB, "5"); err != nil {
		t.Fatalf("expected part size 5 to be accepted: %v", err)
	}
}

// TestGetUploadDirectUrlTTLDefaultsToOneHour verifies the host default when the
// runtime parameter is absent.
func TestGetUploadDirectUrlTTLDefaultsToOneHour(t *testing.T) {
	withRuntimeParamAbsent(t, RuntimeParamKeyUploadDirectUrlTTL)

	svc := New()
	ttl, err := svc.GetUploadDirectUrlTTL(context.Background())
	if err != nil {
		t.Fatalf("get direct url ttl: %v", err)
	}
	if ttl != time.Hour {
		t.Fatalf("expected default direct url ttl 1h, got %s", ttl)
	}
	cfg, err := svc.GetUpload(context.Background())
	if err != nil {
		t.Fatalf("get upload config: %v", err)
	}
	if cfg.DirectUrlTTL != time.Hour {
		t.Fatalf("expected structured direct url ttl 1h, got %s", cfg.DirectUrlTTL)
	}
}

// TestGetUploadDirectUrlTTLPrefersRuntimeParam verifies sys.upload.directUrlTTL
// overrides the host default within the allowed range.
func TestGetUploadDirectUrlTTLPrefersRuntimeParam(t *testing.T) {
	withRuntimeParamValue(t, RuntimeParamKeyUploadDirectUrlTTL, "30m")

	svc := New()
	ttl, err := svc.GetUploadDirectUrlTTL(context.Background())
	if err != nil {
		t.Fatalf("get direct url ttl: %v", err)
	}
	if ttl != 30*time.Minute {
		t.Fatalf("expected runtime direct url ttl 30m, got %s", ttl)
	}
}

// TestValidateUploadDirectUrlTTLConfigValueRejectsOverMax verifies values above
// the one-hour ceiling are rejected at write time.
func TestValidateUploadDirectUrlTTLConfigValueRejectsOverMax(t *testing.T) {
	if err := validateUploadDirectUrlTTLConfigValue(RuntimeParamKeyUploadDirectUrlTTL, "2h"); err == nil {
		t.Fatal("expected validation failure for 2h")
	}
	if err := validateUploadDirectUrlTTLConfigValue(RuntimeParamKeyUploadDirectUrlTTL, "1h"); err != nil {
		t.Fatalf("expected 1h to be accepted: %v", err)
	}
	if err := validateUploadDirectUrlTTLConfigValue(RuntimeParamKeyUploadDirectUrlTTL, "15m"); err != nil {
		t.Fatalf("expected 15m to be accepted: %v", err)
	}
}
