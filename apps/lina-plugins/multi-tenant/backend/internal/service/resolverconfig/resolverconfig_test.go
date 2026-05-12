// This file verifies resolver policy validation for code-owned tenant defaults.

package resolverconfig

import (
	"testing"

	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// TestDefaultConfigDocumentsCodeOwnedTenantDefaults verifies tenant defaults
// are carried by code rather than host config-file values or plugin tables.
func TestDefaultConfigDocumentsCodeOwnedTenantDefaults(t *testing.T) {
	config := defaultConfig()
	if config.RootDomain != shared.DefaultRootDomain {
		t.Fatalf("expected root domain default %q, got %q", shared.DefaultRootDomain, config.RootDomain)
	}
	if config.OnAmbiguous != shared.OnAmbiguousPrompt {
		t.Fatalf("expected ambiguous mode %q, got %q", shared.OnAmbiguousPrompt, config.OnAmbiguous)
	}
	defaultChain := shared.DefaultResolverChain()
	if len(config.Chain) != len(defaultChain) {
		t.Fatalf("expected resolver chain length %d, got %d", len(defaultChain), len(config.Chain))
	}
	for i, expected := range defaultChain {
		if config.Chain[i] != expected {
			t.Fatalf("expected resolver chain item %d to be %q, got %q", i, expected, config.Chain[i])
		}
	}
	defaultReserved := shared.DefaultReservedSubdomains()
	if len(config.ReservedSubdomains) != len(defaultReserved) {
		t.Fatalf("expected reserved subdomain length %d, got %d", len(defaultReserved), len(config.ReservedSubdomains))
	}
	for i, expected := range defaultReserved {
		if config.ReservedSubdomains[i] != expected {
			t.Fatalf("expected reserved subdomain item %d to be %q, got %q", i, expected, config.ReservedSubdomains[i])
		}
	}
}

// TestToResolverConfigKeepsRootDomainDisabled verifies internal callers cannot
// enable subdomain resolution before root-domain configuration is officially
// supported.
func TestToResolverConfigKeepsRootDomainDisabled(t *testing.T) {
	config := ToResolverConfig(&Config{
		Chain:              []string{shared.ResolverSubdomain},
		ReservedSubdomains: []string{"console"},
		RootDomain:         "example.com",
		OnAmbiguous:        shared.OnAmbiguousReject,
	})
	defaultChain := shared.DefaultResolverChain()
	if len(config.Chain) != len(defaultChain) {
		t.Fatalf("expected built-in resolver chain length %d, got %d", len(defaultChain), len(config.Chain))
	}
	for i, expected := range defaultChain {
		if config.Chain[i] != expected {
			t.Fatalf("expected built-in resolver chain item %d to be %q, got %q", i, expected, config.Chain[i])
		}
	}
	defaultReserved := shared.DefaultReservedSubdomains()
	if len(config.ReservedSubdomains) != len(defaultReserved) {
		t.Fatalf("expected built-in reserved length %d, got %d", len(defaultReserved), len(config.ReservedSubdomains))
	}
	for i, expected := range defaultReserved {
		if config.ReservedSubdomains[i] != expected {
			t.Fatalf("expected built-in reserved item %d to be %q, got %q", i, expected, config.ReservedSubdomains[i])
		}
	}
	if config.RootDomain != shared.DefaultRootDomain {
		t.Fatalf("expected resolver root domain to stay disabled, got %q", config.RootDomain)
	}
	if config.OnAmbiguous != shared.OnAmbiguousPrompt {
		t.Fatalf("expected ambiguous mode to stay prompt, got %q", config.OnAmbiguous)
	}
}
