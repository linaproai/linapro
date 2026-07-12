package main

import "testing"

func TestPortFromEnv(t *testing.T) {
	t.Setenv("LINA_CORE_PORT", "18080")
	if got := portFromEnv("LINA_CORE_PORT", 9120); got != 18080 {
		t.Fatalf("primary env: got %d", got)
	}
	t.Setenv("LINA_CORE_PORT", "")
	if got := portFromEnv("LINA_CORE_PORT", 9120); got != 9120 {
		t.Fatalf("fallback: got %d", got)
	}
	t.Setenv("LINA_CORE_PORT", "not-a-port")
	if got := portFromEnv("LINA_CORE_PORT", 9120); got != 9120 {
		t.Fatalf("invalid env should fallback: got %d", got)
	}
	t.Setenv("LINA_VBEN_PORT", "15666")
	if got := portFromEnv("LINA_VBEN_PORT", 5666); got != 15666 {
		t.Fatalf("frontend env: got %d", got)
	}
}
