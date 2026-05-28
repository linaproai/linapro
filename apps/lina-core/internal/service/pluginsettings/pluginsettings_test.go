// This file verifies the pure helpers used by the host plugin settings
// service. Database-backed scenarios are covered separately by integration
// tests; the helpers here are intentionally pure so they stay testable in
// CI without standing up sys_config.

package pluginsettings

import "testing"

// TestBuildKeyComposesNamespacedKey verifies the canonical "<pluginID>.<key>"
// format used to store one setting in sys_config.
func TestBuildKeyComposesNamespacedKey(t *testing.T) {
	got, err := buildKey("linapro-oidc-google", "clientId")
	if err != nil {
		t.Fatalf("buildKey: %v", err)
	}
	if got != "linapro-oidc-google.clientId" {
		t.Fatalf("buildKey = %q, want linapro-oidc-google.clientId", got)
	}
}

// TestBuildKeyTrimsWhitespace verifies leading/trailing whitespace in the
// inputs is rejected via trimming, not silently accepted into the
// persisted key.
func TestBuildKeyTrimsWhitespace(t *testing.T) {
	got, err := buildKey("  linapro-demo  ", "  clientId  ")
	if err != nil {
		t.Fatalf("buildKey: %v", err)
	}
	if got != "linapro-demo.clientId" {
		t.Fatalf("buildKey = %q, want linapro-demo.clientId", got)
	}
}

// TestBuildKeyRejectsEmptyInputs verifies bad inputs are rejected so
// callers cannot accidentally access another plugin's namespace.
func TestBuildKeyRejectsEmptyInputs(t *testing.T) {
	if _, err := buildKey("", "clientId"); err == nil {
		t.Fatal("expected empty pluginID to be rejected")
	}
	if _, err := buildKey("linapro-demo", ""); err == nil {
		t.Fatal("expected empty key to be rejected")
	}
}

// TestBuildKeyRejectsDotInPluginID verifies the helper refuses pluginIDs
// that contain the namespace separator so plugins cannot escape their
// namespace by crafting a key like "linapro-demo.a" with pluginID
// "linapro-demo.a" and key "extra".
func TestBuildKeyRejectsDotInPluginID(t *testing.T) {
	if _, err := buildKey("linapro-demo.bad", "clientId"); err == nil {
		t.Fatal("expected pluginID containing dot to be rejected")
	}
}

// TestMaskSecretShortValue verifies short secrets fold to a fixed
// placeholder so the projection itself does not leak the length.
func TestMaskSecretShortValue(t *testing.T) {
	if got := MaskSecret(""); got != "" {
		t.Fatalf("expected empty mask for empty value, got %q", got)
	}
	if got := MaskSecret("abcdef"); got != "***" {
		t.Fatalf("expected short mask, got %q", got)
	}
}

// TestMaskSecretLongValue verifies longer secrets preserve the first and
// last three characters so operators can tell whether the secret rotated
// without exposing the middle.
func TestMaskSecretLongValue(t *testing.T) {
	if got := MaskSecret("abcdefghijkl"); got != "abc***jkl" {
		t.Fatalf("expected first-and-last preserving mask, got %q", got)
	}
}
