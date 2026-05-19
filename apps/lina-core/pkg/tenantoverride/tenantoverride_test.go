// This file covers tenant override mode normalization and capability checks for
// config and dictionary API projections.

package tenantoverride

import "testing"

// TestNormalizeTenantOverrideModes verifies raw values map to stable published
// override constants.
func TestNormalizeTenantOverrideModes(t *testing.T) {
	cases := map[string]Mode{
		"":                         None,
		"none":                     None,
		" createTenantOverride ":   CreateTenantOverride,
		"createTenantOverride":     CreateTenantOverride,
		"createTenantOverride\t":   CreateTenantOverride,
		"\ncreateTenantOverride\n": CreateTenantOverride,
	}

	for input, expected := range cases {
		if actual := Normalize(input); actual != expected {
			t.Fatalf("Normalize(%q) = %q, want %q", input, actual, expected)
		}
	}
}

// TestCanCreateTenantOverride verifies the helper preserves the narrow action
// semantics instead of treating all supported modes as writable.
func TestCanCreateTenantOverride(t *testing.T) {
	if !CanCreateTenantOverride(CreateTenantOverride) {
		t.Fatal("expected createTenantOverride to allow tenant override creation")
	}
	if CanCreateTenantOverride(None) {
		t.Fatal("expected none to disallow tenant override creation")
	}
}
