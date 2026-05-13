// This file verifies tenant audit metadata resolution for operation logs.

package operlog

import (
	"context"
	"testing"

	plugincontract "lina-core/pkg/pluginservice/contract"
	tenantfilter "lina-core/pkg/pluginservice/tenantfilter"
)

// testBizCtxService returns one configured plugin-visible context snapshot.
type testBizCtxService struct {
	current plugincontract.CurrentContext
}

// Current returns the current plugin-visible business context.
func (s testBizCtxService) Current(context.Context) plugincontract.CurrentContext {
	return s.current
}

// TestResolveAuditTenantContextReadsBizContext verifies tenant metadata comes from bizctx.
func TestResolveAuditTenantContextReadsBizContext(t *testing.T) {
	tenantFilter := newTenantFilterForTest(plugincontract.CurrentContext{
		UserID:   3,
		TenantID: 12,
	})

	actual := resolveAuditTenantContext(context.Background(), tenantFilter, nil, nil, nil, nil)
	if actual.TenantID != 12 || actual.OnBehalfOfTenantID != 0 {
		t.Fatalf("expected tenant 12, got %#v", actual)
	}
	if actual.ActingUserID != 3 || actual.IsImpersonation {
		t.Fatalf("expected regular tenant audit metadata, got %#v", actual)
	}
}

// TestResolveAuditTenantContextReadsImpersonation verifies impersonation metadata comes from bizctx.
func TestResolveAuditTenantContextReadsImpersonation(t *testing.T) {
	tenantFilter := newTenantFilterForTest(plugincontract.CurrentContext{
		UserID:          10,
		TenantID:        12,
		ActingUserID:    3,
		ActingAsTenant:  true,
		IsImpersonation: true,
	})

	actual := resolveAuditTenantContext(context.Background(), tenantFilter, nil, nil, nil, nil)
	if actual.TenantID != 12 || actual.OnBehalfOfTenantID != 12 {
		t.Fatalf("expected impersonation tenant metadata, got %#v", actual)
	}
	if actual.ActingUserID != 3 || !actual.IsImpersonation {
		t.Fatalf("expected impersonation audit metadata, got %#v", actual)
	}
}

// TestResolveAuditTenantContextHonorsExplicitOverrides verifies callers can override context defaults.
func TestResolveAuditTenantContextHonorsExplicitOverrides(t *testing.T) {
	tenantFilter := newTenantFilterForTest(plugincontract.CurrentContext{})
	tenantID := 21
	actingUserID := 4
	onBehalfOfTenantID := 22
	isImpersonation := true

	actual := resolveAuditTenantContext(
		context.Background(),
		tenantFilter,
		&tenantID,
		&actingUserID,
		&onBehalfOfTenantID,
		&isImpersonation,
	)
	if actual.TenantID != 21 || actual.ActingUserID != 4 || actual.OnBehalfOfTenantID != 22 || !actual.IsImpersonation {
		t.Fatalf("expected explicit overrides, got %#v", actual)
	}
}

// newTenantFilterForTest creates an explicitly injected tenant filter service.
func newTenantFilterForTest(current plugincontract.CurrentContext) plugincontract.TenantFilterService {
	return tenantfilter.New(testBizCtxService{current: current}, nil)
}
