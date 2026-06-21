// This file verifies JWT-related configuration loading and runtime overrides.

package config

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

// TestGetJwtUsesDefaultWhenUnset verifies JWT config falls back to its default
// duration when static config and runtime overrides are absent.
func TestGetJwtUsesDefaultWhenUnset(t *testing.T) {
	setTestConfigContent(t, `
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
`)
	withRuntimeParamAbsent(t, RuntimeParamKeyJWTExpire)

	cfg, err := New().GetJwt(context.Background())
	if err != nil {
		t.Fatalf("get jwt config: %v", err)
	}

	if cfg.Expire != 24*time.Hour {
		t.Fatalf("expected default jwt expire to be 24h, got %s", cfg.Expire)
	}
}

// TestGetJwtUsesDurationConfig verifies JWT duration settings come from static
// config when the runtime override is absent.
func TestGetJwtUsesDurationConfig(t *testing.T) {
	setTestConfigContent(t, `
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
jwt:
  secret: "test-secret"
  expire: 36h
`)
	withRuntimeParamAbsent(t, RuntimeParamKeyJWTExpire)

	svc := New()
	cfg, err := svc.GetJwt(context.Background())
	if err != nil {
		t.Fatalf("get jwt config: %v", err)
	}

	if cfg.Expire != 36*time.Hour {
		t.Fatalf("expected jwt expire to be 36h, got %s", cfg.Expire)
	}
	if cfg.Secret != "test-secret" {
		t.Fatalf("expected jwt secret to be test-secret, got %q", cfg.Secret)
	}
	expire, err := svc.GetJwtExpire(context.Background())
	if err != nil {
		t.Fatalf("get jwt expire: %v", err)
	}
	if expire != 36*time.Hour {
		t.Fatalf("expected GetJwtExpire to be 36h, got %s", expire)
	}
	if secret := svc.GetJwtSecret(context.Background()); secret != "test-secret" {
		t.Fatalf("expected GetJwtSecret to be test-secret, got %q", secret)
	}
}

// TestGetJwtPrefersRuntimeParamOverride verifies runtime JWT overrides win over
// static config values.
func TestGetJwtPrefersRuntimeParamOverride(t *testing.T) {
	withRuntimeParamValue(t, RuntimeParamKeyJWTExpire, "12h")

	svc := New()
	cfg, err := svc.GetJwt(context.Background())
	if err != nil {
		t.Fatalf("get jwt config: %v", err)
	}

	if cfg.Expire != 12*time.Hour {
		t.Fatalf("expected runtime param jwt expire to be 12h, got %s", cfg.Expire)
	}
	expire, err := svc.GetJwtExpire(context.Background())
	if err != nil {
		t.Fatalf("get jwt expire: %v", err)
	}
	if expire != 12*time.Hour {
		t.Fatalf("expected runtime getter jwt expire to be 12h, got %s", expire)
	}
}

// TestRuntimeParamParseErrorsReturnError verifies malformed cached runtime
// values are propagated to request-time config readers.
func TestRuntimeParamParseErrorsReturnError(t *testing.T) {
	withCachedRuntimeParamParseError(t, RuntimeParamKeyJWTExpire, gerror.New("bad runtime duration"))

	if _, err := New().GetJwtExpire(context.Background()); err == nil {
		t.Fatal("expected invalid runtime JWT override to return an error")
	}
}
