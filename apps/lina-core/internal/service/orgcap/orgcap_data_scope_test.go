// This file verifies optional organization capability enablement for data
// permission filtering.

package orgcap

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	pkgorgcap "lina-core/pkg/orgcap"
)

// TestServiceRequiresPluginEnabledAndProvider verifies department data-scope
// support is available only when org-center is enabled and a provider exists.
func TestServiceRequiresPluginEnabledAndProvider(t *testing.T) {
	ctx := context.Background()
	pkgorgcap.RegisterProvider(nil)
	t.Cleanup(func() { pkgorgcap.RegisterProvider(nil) })

	svc := New(orgcapEnablementReader{enabled: true})
	if svc.Enabled(ctx) {
		t.Fatal("expected orgcap to stay disabled without provider")
	}

	pkgorgcap.RegisterProvider(&orgcapTestProvider{})
	if !svc.Enabled(ctx) {
		t.Fatal("expected orgcap to be enabled with plugin and provider")
	}

	disabledSvc := New(orgcapEnablementReader{enabled: false})
	if disabledSvc.Enabled(ctx) {
		t.Fatal("expected disabled plugin state to disable orgcap even with provider")
	}
}

// TestServiceBuildUserDeptScopeExistsFallsBackWhenDisabled verifies disabled
// organization capability reports an empty department scope instead of calling
// the registered provider.
func TestServiceBuildUserDeptScopeExistsFallsBackWhenDisabled(t *testing.T) {
	ctx := context.Background()
	provider := &orgcapTestProvider{}
	pkgorgcap.RegisterProvider(provider)
	t.Cleanup(func() { pkgorgcap.RegisterProvider(nil) })

	svc := New(orgcapEnablementReader{enabled: false})
	model, empty, err := svc.BuildUserDeptScopeExists(ctx, "sys_user.id", 21)
	if err != nil {
		t.Fatalf("build disabled orgcap EXISTS: %v", err)
	}
	if model != nil || !empty {
		t.Fatalf("expected disabled orgcap to return empty scope, got model=%v empty=%t", model, empty)
	}
	if provider.existsCalls != 0 {
		t.Fatalf("expected disabled orgcap not to call provider, got %d calls", provider.existsCalls)
	}
}

// orgcapEnablementReader supplies a deterministic plugin enablement state.
type orgcapEnablementReader struct {
	enabled bool
}

// IsEnabled reports the configured enablement for org-center only.
func (r orgcapEnablementReader) IsEnabled(_ context.Context, pluginID string) bool {
	return r.enabled && pluginID == pkgorgcap.ProviderPluginID
}

// orgcapTestProvider records provider calls for orgcap service tests.
type orgcapTestProvider struct {
	existsCalls int
	existsModel *gdb.Model
}

// ListUserDeptAssignments returns no department projections.
func (p orgcapTestProvider) ListUserDeptAssignments(context.Context, []int) (map[int]*pkgorgcap.UserDeptAssignment, error) {
	return map[int]*pkgorgcap.UserDeptAssignment{}, nil
}

// GetUserIDsByDept returns no users.
func (p orgcapTestProvider) GetUserIDsByDept(context.Context, int) ([]int, error) {
	return []int{}, nil
}

// GetAllAssignedUserIDs returns no assigned users.
func (p orgcapTestProvider) GetAllAssignedUserIDs(context.Context) ([]int, error) {
	return []int{}, nil
}

// GetUserDeptInfo returns no department projection.
func (p orgcapTestProvider) GetUserDeptInfo(context.Context, int) (int, string, error) {
	return 0, "", nil
}

// GetUserDeptIDs returns no department IDs.
func (p orgcapTestProvider) GetUserDeptIDs(context.Context, int) ([]int, error) {
	return []int{}, nil
}

// ApplyUserDeptScope returns the original model for test calls.
func (p orgcapTestProvider) ApplyUserDeptScope(_ context.Context, model *gdb.Model, _ string, _ int) (*gdb.Model, bool, error) {
	return model, false, nil
}

// BuildUserDeptScopeExists records a provider-side EXISTS build.
func (p *orgcapTestProvider) BuildUserDeptScopeExists(context.Context, string, int) (*gdb.Model, bool, error) {
	p.existsCalls++
	if p.existsModel == nil {
		return nil, true, nil
	}
	return p.existsModel, false, nil
}

// GetUserPostIDs returns no post IDs.
func (p orgcapTestProvider) GetUserPostIDs(context.Context, int) ([]int, error) {
	return []int{}, nil
}

// ReplaceUserAssignments accepts assignment replacement.
func (p orgcapTestProvider) ReplaceUserAssignments(context.Context, int, *int, []int) error {
	return nil
}

// CleanupUserAssignments accepts assignment cleanup.
func (p orgcapTestProvider) CleanupUserAssignments(context.Context, int) error {
	return nil
}

// UserDeptTree returns no department tree.
func (p orgcapTestProvider) UserDeptTree(context.Context) ([]*pkgorgcap.DeptTreeNode, error) {
	return []*pkgorgcap.DeptTreeNode{}, nil
}

// ListPostOptions returns no post options.
func (p orgcapTestProvider) ListPostOptions(context.Context, *int) ([]*pkgorgcap.PostOption, error) {
	return []*pkgorgcap.PostOption{}, nil
}
