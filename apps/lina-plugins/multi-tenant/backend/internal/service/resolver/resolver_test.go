// This file verifies built-in tenant resolver metadata that does not require a
// live database.

package resolver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"

	"lina-plugin-multi-tenant/backend/internal/service/membership"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// fakeUserTenantLister records default resolver membership lookups for unit tests.
type fakeUserTenantLister struct {
	calls   int
	tenants []*membership.TenantInfo
	err     error
}

// ListUserTenants returns the configured tenant list and records the lookup.
func (f *fakeUserTenantLister) ListUserTenants(_ context.Context, _ int64) ([]*membership.TenantInfo, error) {
	f.calls++
	return f.tenants, f.err
}

// TestBuiltInResolverNames verifies all built-in resolvers publish the stable
// names expected by resolver-chain configuration.
func TestBuiltInResolverNames(t *testing.T) {
	testCases := []struct {
		name     string
		resolver Resolver
		expected string
	}{
		{name: "override", resolver: overrideResolver{}, expected: shared.ResolverOverride},
		{name: "jwt", resolver: jwtResolver{}, expected: shared.ResolverJWT},
		{name: "session", resolver: sessionResolver{}, expected: shared.ResolverSession},
		{name: "header", resolver: headerResolver{}, expected: shared.ResolverHeader},
		{name: "subdomain", resolver: subdomainResolver{}, expected: shared.ResolverSubdomain},
		{name: "default", resolver: defaultResolver{}, expected: shared.ResolverDefault},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if actual := testCase.resolver.Name(); actual != testCase.expected {
				t.Fatalf("expected resolver name %q, got %q", testCase.expected, actual)
			}
		})
	}
}

// TestJWTResolverUsesHostBizContextTenant verifies JWT resolution follows the
// tenant already attached to the host business context by authentication.
func TestJWTResolverUsesHostBizContextTenant(t *testing.T) {
	resolved, matched, err := jwtResolver{}.Resolve(
		context.Background(),
		&ghttp.Request{},
		Identity{UserID: 101, TenantID: 22},
		Config{},
	)
	if err != nil {
		t.Fatalf("resolve jwt tenant failed: %v", err)
	}
	if !matched {
		t.Fatal("expected jwt resolver to match a tenant-bound identity")
	}
	if resolved == nil || resolved.TenantID != 22 || resolved.Source != shared.ResolverJWT {
		t.Fatalf("expected jwt tenant 22, got %#v", resolved)
	}
}

// TestJWTResolverPreservesImpersonation verifies impersonation JWT metadata is
// preserved so platform tenant views bypass ordinary membership checks.
func TestJWTResolverPreservesImpersonation(t *testing.T) {
	resolved, matched, err := jwtResolver{}.Resolve(
		context.Background(),
		&ghttp.Request{},
		Identity{
			UserID:          1,
			TenantID:        22,
			ActingUserID:    1,
			ActingAsTenant:  true,
			IsImpersonation: true,
			IsPlatform:      true,
		},
		Config{},
	)
	if err != nil {
		t.Fatalf("resolve impersonation jwt tenant failed: %v", err)
	}
	if !matched {
		t.Fatal("expected impersonation jwt resolver to match")
	}
	if resolved == nil || resolved.TenantID != 22 || !resolved.ActingAsTenant {
		t.Fatalf("expected acting tenant result for impersonation, got %#v", resolved)
	}
}

// TestValidateMembershipSkipsImpersonationResult verifies trusted platform
// tenant views do not require an ordinary tenant membership row.
func TestValidateMembershipSkipsImpersonationResult(t *testing.T) {
	svc := &serviceImpl{}
	err := svc.validateMembership(
		context.Background(),
		Identity{
			UserID:          1,
			TenantID:        22,
			ActingUserID:    1,
			ActingAsTenant:  true,
			IsImpersonation: true,
		},
		&Result{
			TenantID:       22,
			Source:         shared.ResolverJWT,
			ActingAsTenant: true,
		},
	)
	if err != nil {
		t.Fatalf("expected impersonation membership validation to be skipped, got %v", err)
	}
}

// TestJWTResolverSkipsPlatformTenant verifies platform tokens do not force a
// tenant result and can continue through the configured resolver chain.
func TestJWTResolverSkipsPlatformTenant(t *testing.T) {
	resolved, matched, err := jwtResolver{}.Resolve(
		context.Background(),
		&ghttp.Request{},
		Identity{UserID: 1, TenantID: shared.PlatformTenantID},
		Config{},
	)
	if err != nil {
		t.Fatalf("resolve platform jwt tenant failed: %v", err)
	}
	if matched || resolved != nil {
		t.Fatalf("expected platform jwt resolver to skip, got matched=%t result=%#v", matched, resolved)
	}
}

// TestDefaultResolverKeepsPlatformIdentityAtPlatformTenant verifies platform
// users do not automatically fall into a tenant through membership fallback.
func TestDefaultResolverKeepsPlatformIdentityAtPlatformTenant(t *testing.T) {
	memberships := &fakeUserTenantLister{
		tenants: []*membership.TenantInfo{{Id: 22, Code: "acme", Name: "Acme", Status: string(shared.TenantStatusActive)}},
	}

	resolved, matched, err := defaultResolver{membershipSvc: memberships}.Resolve(
		context.Background(),
		&ghttp.Request{},
		Identity{UserID: 1, TenantID: shared.PlatformTenantID, IsPlatform: true},
		Config{},
	)
	if err != nil {
		t.Fatalf("resolve platform default tenant failed: %v", err)
	}
	if !matched {
		t.Fatal("expected default resolver to match platform identity")
	}
	if resolved == nil || resolved.TenantID != shared.PlatformTenantID || resolved.Source != shared.ResolverDefault {
		t.Fatalf("expected platform default tenant, got %#v", resolved)
	}
	if memberships.calls != 0 {
		t.Fatalf("expected platform identity to skip membership lookup, got %d calls", memberships.calls)
	}
}

// TestDefaultResolverUsesMembershipForNonPlatformIdentity verifies ordinary
// users keep resolving to their first enabled tenant membership.
func TestDefaultResolverUsesMembershipForNonPlatformIdentity(t *testing.T) {
	memberships := &fakeUserTenantLister{
		tenants: []*membership.TenantInfo{
			{Id: 22, Code: "acme", Name: "Acme", Status: string(shared.TenantStatusActive)},
			{Id: 33, Code: "beta", Name: "Beta", Status: string(shared.TenantStatusActive)},
		},
	}

	resolved, matched, err := defaultResolver{membershipSvc: memberships}.Resolve(
		context.Background(),
		&ghttp.Request{},
		Identity{UserID: 101, TenantID: shared.PlatformTenantID},
		Config{},
	)
	if err != nil {
		t.Fatalf("resolve tenant membership failed: %v", err)
	}
	if !matched {
		t.Fatal("expected default resolver to match non-platform identity")
	}
	if resolved == nil || resolved.TenantID != 22 || resolved.Source != shared.ResolverDefault {
		t.Fatalf("expected first membership tenant 22, got %#v", resolved)
	}
	if memberships.calls != 1 {
		t.Fatalf("expected one membership lookup, got %d calls", memberships.calls)
	}
}

// TestOverrideResolverRejectsNonPlatformIdentity verifies tenant override is
// restricted to platform administrators before the header value is trusted.
func TestOverrideResolverRejectsNonPlatformIdentity(t *testing.T) {
	request := &ghttp.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set("X-Tenant-Override", "22")

	resolved, matched, err := overrideResolver{}.Resolve(
		context.Background(),
		request,
		Identity{UserID: 101, TenantID: 22},
		Config{},
	)
	if !matched {
		t.Fatal("expected override resolver to match the override header")
	}
	if resolved != nil {
		t.Fatalf("expected no result for rejected override, got %#v", resolved)
	}
	if !bizerr.Is(err, CodePlatformPermissionRequired) {
		t.Fatalf("expected platform permission error, got %v", err)
	}
}

// TestOverrideResolverWrapsInvalidTenantID verifies malformed override headers
// return a stable business error instead of leaking parser details.
func TestOverrideResolverWrapsInvalidTenantID(t *testing.T) {
	request := &ghttp.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set("X-Tenant-Override", "not-a-tenant-id")

	resolved, matched, err := overrideResolver{}.Resolve(
		context.Background(),
		request,
		Identity{UserID: 1, IsPlatform: true},
		Config{},
	)
	if !matched {
		t.Fatal("expected override resolver to match the override header")
	}
	if resolved != nil {
		t.Fatalf("expected no result for invalid override, got %#v", resolved)
	}
	if !bizerr.Is(err, CodeTenantOverrideInvalid) {
		t.Fatalf("expected invalid override error, got %v", err)
	}
}

// TestSubdomainLabelRespectsRootDomain verifies subdomain parsing uses the configured root domain.
func TestSubdomainLabelRespectsRootDomain(t *testing.T) {
	testCases := []struct {
		name       string
		host       string
		rootDomain string
		expected   string
	}{
		{name: "root domain", host: "acme.example.com", rootDomain: "example.com", expected: "acme"},
		{name: "with port", host: "acme.example.com:5666", rootDomain: "example.com", expected: "acme"},
		{name: "wrong root", host: "acme.other.com", rootDomain: "example.com", expected: ""},
		{name: "nested root", host: "eu.acme.example.com", rootDomain: "example.com", expected: ""},
		{name: "no root disables resolver", host: "acme.localhost", rootDomain: "", expected: ""},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if actual := subdomainLabel(testCase.host, testCase.rootDomain); actual != testCase.expected {
				t.Fatalf("expected subdomain label %q, got %q", testCase.expected, actual)
			}
		})
	}
}

// TestReservedSubdomainConfig verifies reserved labels can be supplied by resolver policy.
func TestReservedSubdomainConfig(t *testing.T) {
	if !isReservedSubdomain("www", nil) {
		t.Fatal("expected built-in www label to be reserved")
	}
	if isReservedSubdomain("www", []string{"console"}) {
		t.Fatal("expected configured reservations to replace built-in labels")
	}
	if !isReservedSubdomain("console", []string{"console"}) {
		t.Fatal("expected configured console label to be reserved")
	}
}

// TestFindTenantByCodeTreatsMissingTenantAsNoMatch verifies resolver probes
// can continue when a host or header label does not map to any tenant.
func TestFindTenantByCodeTreatsMissingTenantAsNoMatch(t *testing.T) {
	ctx := context.Background()
	configureResolverTestDB(t, ctx)
	prepareResolverTenantSchema(t, ctx)

	resolved, matched, err := findTenantByCode(ctx, "tenant-code-that-does-not-exist", shared.ResolverSubdomain)
	if err != nil {
		t.Fatalf("expected missing tenant code to be treated as no match, got error: %v", err)
	}
	if matched || resolved != nil {
		t.Fatalf("expected no resolver match for missing tenant code, got matched=%t result=%#v", matched, resolved)
	}
}

// prepareResolverTenantSchema creates the minimal tenant table needed by
// resolver lookup tests.
func prepareResolverTenantSchema(t *testing.T, ctx context.Context) {
	t.Helper()

	statement := `CREATE TABLE IF NOT EXISTS plugin_multi_tenant_tenant (
		"id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		"code" VARCHAR(64) NOT NULL,
		"name" VARCHAR(128) NOT NULL,
		"status" VARCHAR(32) NOT NULL DEFAULT 'active',
		"remark" VARCHAR(512) NOT NULL DEFAULT '',
		"created_by" BIGINT NOT NULL DEFAULT 0,
		"updated_by" BIGINT NOT NULL DEFAULT 0,
		"created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"deleted_at" TIMESTAMP,
		CONSTRAINT uk_plugin_multi_tenant_tenant_code UNIQUE ("code")
	)`
	if _, err := g.DB().Exec(ctx, statement); err != nil {
		t.Fatalf("prepare resolver tenant schema failed: %v", err)
	}
}

// configureResolverTestDB points the package test at the local PostgreSQL
// database initialized by the repository test workflow.
func configureResolverTestDB(t *testing.T, ctx context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"}},
	}); err != nil {
		t.Fatalf("configure resolver test database failed: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if err := db.Close(ctx); err != nil {
			t.Errorf("close resolver test database failed: %v", err)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore resolver test database config failed: %v", err)
		}
	})
}
