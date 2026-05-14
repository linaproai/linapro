// This file provides shared helpers for Wasm host-service tests.

package wasm

import (
	"testing"
)

// assertPanic verifies the supplied function panics with the expected message.
func assertPanic(t *testing.T, expected string, fn func()) {
	t.Helper()

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("expected panic %q", expected)
		}
		if recovered != expected {
			t.Fatalf("expected panic %q, got %#v", expected, recovered)
		}
	}()

	fn()
}
