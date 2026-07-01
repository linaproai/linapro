// This file verifies organization capability fallback behavior when no
// provider is active. These checks protect host services from turning optional
// organization features into hard runtime dependencies.

package orgspi

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap"
)

func TestDisabledOrganizationCapabilityReturnsNeutralValues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := New(nil, nil, nil)

	if svc.Available(ctx) {
		t.Fatal("expected disabled organization capability")
	}
	if svc.Available(ctx) {
		t.Fatal("expected unavailable organization provider")
	}
	if status := svc.Status(ctx); status.Available || status.ActiveProvider != "" {
		t.Fatalf("expected unavailable status without active provider, got %#v", status)
	}

	assignments, err := svc.Assignment().BatchListByUsers(ctx, []int{1, 2})
	if err != nil {
		t.Fatalf("list department assignments returned error: %v", err)
	}
	if len(assignments) != 0 {
		t.Fatalf("expected empty assignment projection, got %#v", assignments)
	}

	deptID, deptName, err := svc.Assignment().GetUserDeptInfo(ctx, 1)
	if err != nil {
		t.Fatalf("get user dept info returned error: %v", err)
	}
	if deptID != 0 || deptName != "" {
		t.Fatalf("expected zero department projection, got id=%d name=%q", deptID, deptName)
	}

	if ids, err := svc.Assignment().GetUserDeptIDs(ctx, 1); err != nil || len(ids) != 0 {
		t.Fatalf("expected empty department IDs without error, got ids=%#v err=%v", ids, err)
	}
	if ids, err := svc.Assignment().GetUserPostIDs(ctx, 1); err != nil || len(ids) != 0 {
		t.Fatalf("expected empty post IDs without error, got ids=%#v err=%v", ids, err)
	}
}

func TestDisabledOrganizationCapabilityKeepsHostInternalOperationsSafe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := New(nil, nil, nil)

	model, applied, err := svc.Scope().ApplyUserDeptScope(ctx, nil, "user_id", 1)
	if err != nil {
		t.Fatalf("apply user department scope returned error: %v", err)
	}
	if model != nil {
		t.Fatalf("expected nil model to remain unchanged, got %#v", model)
	}
	if !applied {
		t.Fatal("expected disabled organization scope to report permissive fallback")
	}

	exists, applied, err := svc.Scope().BuildUserDeptScopeExists(ctx, "user_id", 1)
	if err != nil {
		t.Fatalf("build user department scope exists returned error: %v", err)
	}
	if exists != nil {
		t.Fatalf("expected nil exists model when provider is unavailable, got %#v", exists)
	}
	if !applied {
		t.Fatal("expected disabled organization exists scope to report permissive fallback")
	}

	filtered, applied, err := svc.Scope().ApplyUserDeptFilter(ctx, nil, "user_id", 10)
	if err != nil {
		t.Fatalf("apply user department filter returned error: %v", err)
	}
	if filtered != nil || applied {
		t.Fatalf("expected disabled organization filter to remain unchanged, got model=%#v applied=%v", filtered, applied)
	}

	unassigned, applied, err := svc.Scope().ApplyUserDeptUnassignedFilter(ctx, nil, "user_id")
	if err != nil {
		t.Fatalf("apply user department unassigned filter returned error: %v", err)
	}
	if unassigned != nil || applied {
		t.Fatalf("expected disabled organization unassigned filter to remain unchanged, got model=%#v applied=%v", unassigned, applied)
	}

	if err = svc.Assignment().ReplaceByUser(ctx, 1, nil, []int{2}); err != nil {
		t.Fatalf("replace user assignments should be a no-op when disabled: %v", err)
	}
	if err = svc.Assignment().CleanupByUser(ctx, 1); err != nil {
		t.Fatalf("cleanup user assignments should be a no-op when disabled: %v", err)
	}
	if tree, err := svc.Workspace().UserDeptTree(ctx); err != nil || len(tree) != 0 {
		t.Fatalf("expected empty department tree without error, got tree=%#v err=%v", tree, err)
	}
	if posts, err := svc.Workspace().ListPostOptions(ctx, nil); err != nil || len(posts) != 0 {
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

	runtime := orgConflictRuntime{enabled: enabled}
	svc := New(manager, runtime, runtime.ProviderEnv)
	_, err := svc.Department().ListTree(ctx, orgcap.DeptTreeInput{MaxNodes: 10})
	if !bizerr.Is(err, capmodel.CodeCapabilityProviderConflict) {
		t.Fatalf("expected provider conflict error, got %v", err)
	}
	status := svc.Status(ctx)
	if status.Available || status.Reason != "provider_conflict" {
		t.Fatalf("expected unavailable provider conflict status, got %#v", status)
	}
}

func TestOrganizationCapabilityDerivesConvenienceReadsFromBatchProvider(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewManager()
	const pluginID = "org-derived-provider"
	provider := &derivedOrgProvider{}
	if err := manager.RegisterFactory(pluginID, func(context.Context, ProviderEnv) (Provider, error) {
		return provider, nil
	}); err != nil {
		t.Fatalf("register organization provider: %v", err)
	}
	runtime := orgConflictRuntime{enabled: map[string]bool{pluginID: true}}
	svc := New(manager, runtime, runtime.ProviderEnv)

	assignments, err := svc.Assignment().BatchListByUsers(ctx, []int{7})
	if err != nil {
		t.Fatalf("batch list assignments: %v", err)
	}
	if assignments[7] == nil || assignments[7].DeptID != 17 || assignments[7].DeptName != "Dept-7" {
		t.Fatalf("expected assignment derived from profile, got %#v", assignments)
	}
	deptID, deptName, err := svc.Assignment().GetUserDeptInfo(ctx, 7)
	if err != nil || deptID != 17 || deptName != "Dept-7" {
		t.Fatalf("expected department info from profile, got id=%d name=%q err=%v", deptID, deptName, err)
	}
	if deptIDs, err := svc.Assignment().GetUserDeptIDs(ctx, 7); err != nil || len(deptIDs) != 1 || deptIDs[0] != 17 {
		t.Fatalf("expected department IDs from profile, got ids=%#v err=%v", deptIDs, err)
	}
	if postIDs, err := svc.Assignment().GetUserPostIDs(ctx, 7); err != nil || len(postIDs) != 1 || postIDs[0] != 107 {
		t.Fatalf("expected post IDs from profile, got ids=%#v err=%v", postIDs, err)
	}
	if err := svc.Department().EnsureVisible(ctx, []int{17}); err != nil {
		t.Fatalf("expected visible department to pass derived visibility check: %v", err)
	}
	if err := svc.Post().EnsureVisible(ctx, []int{107}); err != nil {
		t.Fatalf("expected visible post to pass derived visibility check: %v", err)
	}
	options, err := svc.Workspace().ListPostOptions(ctx, nil)
	if err != nil {
		t.Fatalf("list post options: %v", err)
	}
	if len(options) != 1 || options[0].PostID != 107 || options[0].PostName != "Post-7" {
		t.Fatalf("expected post options derived from post list, got %#v", options)
	}
	if provider.lastPostListStatus == nil || *provider.lastPostListStatus != 1 {
		t.Fatalf("expected post options to default to enabled status, got %#v", provider.lastPostListStatus)
	}
	tree, err := svc.Workspace().UserDeptTree(ctx)
	if err != nil {
		t.Fatalf("user department tree: %v", err)
	}
	if len(tree) != 1 || tree[0].Id != 17 {
		t.Fatalf("expected workspace tree from bounded department tree, got %#v", tree)
	}
}

type orgConflictRuntime struct {
	enabled map[string]bool
}

func (s orgConflictRuntime) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return s.enabled[pluginID]
}

func (s orgConflictRuntime) ProviderEnv(_ context.Context, pluginID string) ProviderEnv {
	return ProviderEnv{PluginID: pluginID}
}

type derivedOrgProvider struct {
	lastPostListStatus *int
}

func (derivedOrgProvider) BatchGetDepartments(
	_ context.Context,
	deptIDs []int,
) (*capmodel.BatchResult[*orgcap.DeptInfo, int], error) {
	result := &capmodel.BatchResult[*orgcap.DeptInfo, int]{Items: map[int]*orgcap.DeptInfo{}, MissingIDs: []int{}}
	for _, deptID := range deptIDs {
		if deptID == 17 {
			result.Items[deptID] = &orgcap.DeptInfo{DeptID: deptID, DeptName: "Dept-7"}
			continue
		}
		result.MissingIDs = append(result.MissingIDs, deptID)
	}
	return result, nil
}

func (derivedOrgProvider) SearchDepartments(
	context.Context,
	orgcap.DeptListInput,
) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	return &capmodel.PageResult[*orgcap.DeptInfo]{Items: []*orgcap.DeptInfo{{DeptID: 17, DeptName: "Dept-7"}}, Total: 1}, nil
}

func (derivedOrgProvider) ListDeptTree(context.Context, orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error) {
	return &orgcap.DeptTreeResult{Items: []*orgcap.DeptTreeNode{{Id: 17, Label: "Dept-7"}}, Total: 1}, nil
}

func (derivedOrgProvider) CreateDepartment(context.Context, orgcap.DeptCreateInput) (int, error) {
	return 17, nil
}

func (derivedOrgProvider) UpdateDepartment(context.Context, orgcap.DeptUpdateInput) error {
	return nil
}

func (derivedOrgProvider) DeleteDepartment(context.Context, int) error {
	return nil
}

func (derivedOrgProvider) BatchGetPosts(
	_ context.Context,
	postIDs []int,
) (*capmodel.BatchResult[*orgcap.PostInfo, int], error) {
	result := &capmodel.BatchResult[*orgcap.PostInfo, int]{Items: map[int]*orgcap.PostInfo{}, MissingIDs: []int{}}
	for _, postID := range postIDs {
		if postID == 107 {
			result.Items[postID] = &orgcap.PostInfo{PostID: postID, PostName: "Post-7"}
			continue
		}
		result.MissingIDs = append(result.MissingIDs, postID)
	}
	return result, nil
}

func (p *derivedOrgProvider) ListPosts(_ context.Context, input orgcap.PostListInput) (*capmodel.PageResult[*orgcap.PostInfo], error) {
	p.lastPostListStatus = input.Status
	return &capmodel.PageResult[*orgcap.PostInfo]{Items: []*orgcap.PostInfo{{PostID: 107, PostName: "Post-7"}}, Total: 1}, nil
}

func (derivedOrgProvider) CreatePost(context.Context, orgcap.PostCreateInput) (int, error) {
	return 107, nil
}

func (derivedOrgProvider) UpdatePost(context.Context, orgcap.PostUpdateInput) error {
	return nil
}

func (derivedOrgProvider) DeletePost(context.Context, int) error {
	return nil
}

func (derivedOrgProvider) BatchGetUserOrgProfiles(
	_ context.Context,
	userIDs []int,
) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error) {
	result := &capmodel.BatchResult[*orgcap.UserOrgProfile, int]{Items: map[int]*orgcap.UserOrgProfile{}, MissingIDs: []int{}}
	for _, userID := range userIDs {
		result.Items[userID] = &orgcap.UserOrgProfile{
			UserID:    userID,
			DeptID:    userID + 10,
			DeptName:  "Dept-7",
			PostIDs:   []int{userID + 100},
			PostNames: []string{"Post-7"},
		}
	}
	return result, nil
}

func (derivedOrgProvider) ReplaceUserAssignments(context.Context, int, *int, []int) error {
	return nil
}

func (derivedOrgProvider) CleanupUserAssignments(context.Context, int) error {
	return nil
}

func (derivedOrgProvider) BuildUserDeptScopeExists(context.Context, string, int) (*gdb.Model, bool, error) {
	return nil, true, nil
}

func (derivedOrgProvider) ApplyUserDeptFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
	_ int,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

func (derivedOrgProvider) ApplyUserDeptUnassignedFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
) (*gdb.Model, bool, error) {
	return model, false, nil
}
