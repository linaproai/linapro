// This file verifies organization capability fallback behavior when no
// provider is active. These checks protect host services from turning optional
// organization features into hard runtime dependencies.

package orgspi

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap"
)

func TestDisabledOrganizationCapabilityReturnsNeutralValues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := New(nil, nil)

	if svc.Available(ctx) {
		t.Fatal("expected disabled organization capability")
	}
	if svc.Available(ctx) {
		t.Fatal("expected unavailable organization provider")
	}
	if status := svc.Status(ctx); status.Available || status.ActiveProvider != "" {
		t.Fatalf("expected unavailable status without active provider, got %#v", status)
	}

	assignments, err := svc.ListUserDeptAssignments(ctx, []int{1, 2})
	if err != nil {
		t.Fatalf("list department assignments returned error: %v", err)
	}
	if len(assignments) != 0 {
		t.Fatalf("expected empty assignment projection, got %#v", assignments)
	}

	deptID, deptName, err := svc.GetUserDeptInfo(ctx, 1)
	if err != nil {
		t.Fatalf("get user dept info returned error: %v", err)
	}
	if deptID != 0 || deptName != "" {
		t.Fatalf("expected zero department projection, got id=%d name=%q", deptID, deptName)
	}

	if name, err := svc.GetUserDeptName(ctx, 1); err != nil || name != "" {
		t.Fatalf("expected empty department name without error, got name=%q err=%v", name, err)
	}
	if ids, err := svc.GetUserDeptIDs(ctx, 1); err != nil || len(ids) != 0 {
		t.Fatalf("expected empty department IDs without error, got ids=%#v err=%v", ids, err)
	}
	if ids, err := svc.GetUserPostIDs(ctx, 1); err != nil || len(ids) != 0 {
		t.Fatalf("expected empty post IDs without error, got ids=%#v err=%v", ids, err)
	}
}

func TestDisabledOrganizationCapabilityKeepsHostInternalOperationsSafe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := New(nil, nil)

	model, applied, err := svc.ApplyUserDeptScope(ctx, nil, "user_id", 1)
	if err != nil {
		t.Fatalf("apply user department scope returned error: %v", err)
	}
	if model != nil {
		t.Fatalf("expected nil model to remain unchanged, got %#v", model)
	}
	if !applied {
		t.Fatal("expected disabled organization scope to report permissive fallback")
	}

	exists, applied, err := svc.BuildUserDeptScopeExists(ctx, "user_id", 1)
	if err != nil {
		t.Fatalf("build user department scope exists returned error: %v", err)
	}
	if exists != nil {
		t.Fatalf("expected nil exists model when provider is unavailable, got %#v", exists)
	}
	if !applied {
		t.Fatal("expected disabled organization exists scope to report permissive fallback")
	}

	filtered, applied, err := svc.ApplyUserDeptFilter(ctx, nil, "user_id", 10)
	if err != nil {
		t.Fatalf("apply user department filter returned error: %v", err)
	}
	if filtered != nil || applied {
		t.Fatalf("expected disabled organization filter to remain unchanged, got model=%#v applied=%v", filtered, applied)
	}

	unassigned, applied, err := svc.ApplyUserDeptUnassignedFilter(ctx, nil, "user_id")
	if err != nil {
		t.Fatalf("apply user department unassigned filter returned error: %v", err)
	}
	if unassigned != nil || applied {
		t.Fatalf("expected disabled organization unassigned filter to remain unchanged, got model=%#v applied=%v", unassigned, applied)
	}

	if err = svc.ReplaceUserAssignments(ctx, 1, nil, []int{2}); err != nil {
		t.Fatalf("replace user assignments should be a no-op when disabled: %v", err)
	}
	if err = svc.CleanupUserAssignments(ctx, 1); err != nil {
		t.Fatalf("cleanup user assignments should be a no-op when disabled: %v", err)
	}
	if tree, err := svc.UserDeptTree(ctx); err != nil || len(tree) != 0 {
		t.Fatalf("expected empty department tree without error, got tree=%#v err=%v", tree, err)
	}
	if posts, err := svc.ListPostOptions(ctx, nil); err != nil || len(posts) != 0 {
		t.Fatalf("expected empty post options without error, got posts=%#v err=%v", posts, err)
	}
}

func TestOrganizationCapabilityRejectsMultipleEnabledProviders(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewManager()
	enabled := map[string]bool{
		"org-provider-conflict-a": true,
		"org-provider-conflict-b": true,
	}
	for pluginID := range enabled {
		pluginID := pluginID
		if err := manager.RegisterFactory(pluginID, func(context.Context, ProviderEnv) (Provider, error) {
			return nil, nil
		}); err != nil {
			t.Fatalf("register organization provider %s: %v", pluginID, err)
		}
	}

	svc := New(manager, orgConflictRuntime{enabled: enabled})
	_, err := svc.ListDeptTree(ctx, orgcap.DeptTreeInput{MaxNodes: 10})
	if !bizerr.Is(err, capmodel.CodeCapabilityProviderConflict) {
		t.Fatalf("expected provider conflict error, got %v", err)
	}
	status := svc.Status(ctx)
	if status.Available || status.Reason != "provider_conflict" {
		t.Fatalf("expected unavailable provider conflict status, got %#v", status)
	}
}

type orgConflictRuntime struct {
	enabled map[string]bool
}

func (s orgConflictRuntime) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return s.enabled[pluginID]
}

func (s orgConflictRuntime) OrgProviderEnv(pluginID string) ProviderEnv {
	return ProviderEnv{PluginID: pluginID}
}
