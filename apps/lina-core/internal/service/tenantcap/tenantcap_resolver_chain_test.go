// This file verifies tenant resolver-chain dispatch behavior.

package tenantcap

import (
	"context"
	"errors"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"

	pkgtenantcap "lina-core/pkg/tenantcap"
)

// chainTestResolver provides deterministic resolver outcomes for chain tests.
type chainTestResolver struct {
	name   pkgtenantcap.ResolverName
	match  bool
	tenant pkgtenantcap.TenantID
	calls  int
	err    error
}

// Name returns the resolver name.
func (r *chainTestResolver) Name() pkgtenantcap.ResolverName {
	return r.name
}

// Resolve returns the configured resolver result.
func (r *chainTestResolver) Resolve(context.Context, *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	return &pkgtenantcap.ResolverResult{TenantID: r.tenant, Matched: r.match}, nil
}

// TestResolverChainUsesConfiguredOrder verifies the first matching configured resolver wins.
func TestResolverChainUsesConfiguredOrder(t *testing.T) {
	header := &chainTestResolver{name: "header", match: true, tenant: 2}
	jwt := &chainTestResolver{name: "jwt", match: true, tenant: 1}
	chain := NewResolverChain([]pkgtenantcap.ResolverName{"jwt", "header"}, header, jwt)

	result, err := chain.Resolve(context.Background(), nil)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if result == nil || result.TenantID != 1 {
		t.Fatalf("expected jwt tenant 1, got %#v", result)
	}
	if jwt.calls != 1 || header.calls != 0 {
		t.Fatalf("expected jwt only, got jwt=%d header=%d", jwt.calls, header.calls)
	}
}

// TestResolverChainSkipsNonMatchingResolvers verifies the chain continues after empty results.
func TestResolverChainSkipsNonMatchingResolvers(t *testing.T) {
	override := &chainTestResolver{name: "override", match: false}
	defaultResolver := &chainTestResolver{name: "default", match: true, tenant: 9}
	chain := NewResolverChain([]pkgtenantcap.ResolverName{"override", "default"}, override, defaultResolver)

	result, err := chain.Resolve(context.Background(), nil)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if result == nil || result.TenantID != 9 {
		t.Fatalf("expected default tenant 9, got %#v", result)
	}
	if override.calls != 1 || defaultResolver.calls != 1 {
		t.Fatalf("expected both resolvers, got override=%d default=%d", override.calls, defaultResolver.calls)
	}
}

// TestResolverChainNormalizesOrder verifies duplicate and blank configured
// resolver names do not change first-match semantics.
func TestResolverChainNormalizesOrder(t *testing.T) {
	tenant := &chainTestResolver{name: "tenant", match: true, tenant: 20}
	fallback := &chainTestResolver{name: "fallback", match: true, tenant: 30}
	chain := NewResolverChain(
		[]pkgtenantcap.ResolverName{" ", "tenant", "tenant", "fallback"},
		fallback,
		tenant,
	)

	result, err := chain.Resolve(context.Background(), nil)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if result == nil || result.TenantID != 20 {
		t.Fatalf("expected normalized first tenant result, got %#v", result)
	}
	if tenant.calls != 1 || fallback.calls != 0 {
		t.Fatalf("expected duplicate tenant to run once, got tenant=%d fallback=%d", tenant.calls, fallback.calls)
	}
}

// TestResolverChainReturnsNilWhenNoResolverMatches verifies provider adapters
// can apply their own fallback policy after an exhausted chain.
func TestResolverChainReturnsNilWhenNoResolverMatches(t *testing.T) {
	header := &chainTestResolver{name: "header", match: false}
	chain := NewResolverChain([]pkgtenantcap.ResolverName{"header", "missing"}, header)

	result, err := chain.Resolve(context.Background(), nil)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result after no resolver matched, got %#v", result)
	}
	if header.calls != 1 {
		t.Fatalf("expected header resolver to be called once, got %d", header.calls)
	}
}

// TestResolverChainStopsOnResolverError verifies resolver errors are surfaced
// without continuing to later fallbacks.
func TestResolverChainStopsOnResolverError(t *testing.T) {
	expectedErr := errors.New("resolver failed")
	header := &chainTestResolver{name: "header", err: expectedErr}
	fallback := &chainTestResolver{name: "fallback", match: true, tenant: 4}
	chain := NewResolverChain([]pkgtenantcap.ResolverName{"header", "fallback"}, header, fallback)

	result, err := chain.Resolve(context.Background(), nil)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected resolver error, got result=%#v err=%v", result, err)
	}
	if fallback.calls != 0 {
		t.Fatalf("expected fallback not to run after error, got %d calls", fallback.calls)
	}
}
