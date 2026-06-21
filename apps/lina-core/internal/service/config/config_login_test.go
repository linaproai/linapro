// This file verifies runtime login parameters managed through sys_config.

package config

import (
	"context"
	"testing"
)

// TestGetLoginUsesRuntimeBlacklist verifies runtime blacklist rules are parsed
// once and reused by both config objects and convenience getters.
func TestGetLoginUsesRuntimeBlacklist(t *testing.T) {
	withRuntimeParamValue(t, RuntimeParamKeyLoginBlackIPList, "127.0.0.1;10.0.0.0/8")

	svc := New()
	cfg, err := svc.GetLogin(context.Background())
	if err != nil {
		t.Fatalf("get runtime login config: %v", err)
	}

	if !cfg.IsBlacklisted("127.0.0.1") {
		t.Fatal("expected 127.0.0.1 to be blacklisted")
	}
	if !cfg.IsBlacklisted("10.1.2.3") {
		t.Fatal("expected 10.1.2.3 to match blacklisted CIDR")
	}
	if cfg.IsBlacklisted("192.168.1.10") {
		t.Fatal("expected 192.168.1.10 not to be blacklisted")
	}
	blacklisted, err := svc.IsLoginIPBlacklisted(context.Background(), "10.1.2.3")
	if err != nil {
		t.Fatalf("check blacklisted runtime IP: %v", err)
	}
	if !blacklisted {
		t.Fatal("expected runtime blacklist getter to match 10.1.2.3")
	}
	blacklisted, err = svc.IsLoginIPBlacklisted(context.Background(), "192.168.1.10")
	if err != nil {
		t.Fatalf("check allowed runtime IP: %v", err)
	}
	if blacklisted {
		t.Fatal("expected runtime blacklist getter not to match 192.168.1.10")
	}
}
