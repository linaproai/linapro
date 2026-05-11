// This file verifies resolver policy validation for code-owned tenant defaults.

package resolverconfig

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
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

// TestUpdateAcceptsOnlyBuiltInPolicy verifies runtime resolver mutations are rejected.
func TestUpdateAcceptsOnlyBuiltInPolicy(t *testing.T) {
	svc := New()
	ctx := context.Background()
	defaults := defaultConfig()

	if err := svc.Update(ctx, UpdateInput{
		Chain:              defaults.Chain,
		ReservedSubdomains: defaults.ReservedSubdomains,
		RootDomain:         defaults.RootDomain,
		OnAmbiguous:        defaults.OnAmbiguous,
	}); err != nil {
		t.Fatalf("expected built-in policy write to be a no-op: %v", err)
	}

	testCases := []UpdateInput{
		{Chain: []string{shared.ResolverJWT, shared.ResolverDefault}, ReservedSubdomains: defaults.ReservedSubdomains, OnAmbiguous: defaults.OnAmbiguous},
		{Chain: defaults.Chain, ReservedSubdomains: []string{"console"}, OnAmbiguous: defaults.OnAmbiguous},
		{Chain: defaults.Chain, ReservedSubdomains: defaults.ReservedSubdomains, RootDomain: "example.com", OnAmbiguous: defaults.OnAmbiguous},
		{Chain: defaults.Chain, ReservedSubdomains: defaults.ReservedSubdomains, OnAmbiguous: shared.OnAmbiguousReject},
	}
	for _, testCase := range testCases {
		if err := svc.Update(ctx, testCase); !bizerr.Is(err, CodeResolverConfigInvalid) {
			t.Fatalf("expected invalid resolver policy error, got %v", err)
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
