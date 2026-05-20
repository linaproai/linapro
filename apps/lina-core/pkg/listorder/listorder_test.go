// This file verifies list-order normalization behavior without importing any
// API or database packages, keeping the component contract self-contained.

package listorder

import "testing"

// TestNormalizeAcceptsSupportedDirections verifies raw caller input is trimmed
// and folded into the canonical list-order constants.
func TestNormalizeAcceptsSupportedDirections(t *testing.T) {
	cases := map[string]Direction{
		"asc":    ASC,
		" ASC ":  ASC,
		"desc":   DESC,
		" DESC ": DESC,
	}

	for input, expected := range cases {
		if actual := Normalize(input); actual != expected {
			t.Fatalf("Normalize(%q) = %q, want %q", input, actual, expected)
		}
	}
}

// TestNormalizeOrDefaultFallsBack verifies unsupported values do not leak into
// query adapters that expect one of the stable directions.
func TestNormalizeOrDefaultFallsBack(t *testing.T) {
	if actual := NormalizeOrDefault("sideways", ASC); actual != ASC {
		t.Fatalf("expected explicit fallback for unsupported input, got %q", actual)
	}
	if actual := NormalizeOrDefault("", Direction("bad")); actual != DESC {
		t.Fatalf("expected package default for invalid fallback, got %q", actual)
	}
}
