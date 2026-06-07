// orgcap_impl.go implements optional organization-capability delegation. It
// checks source-plugin enablement before forwarding department, post, and
// data-scope operations, returning neutral values when the provider is absent
// so host services can degrade without hard dependencies on organization data.

package orgcap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"

	"github.com/gogf/gf/v2/database/gdb"
)

// IsProviderEnabled always returns false.
func (noopProviderRuntime) IsProviderEnabled(_ context.Context, _ string) bool {
	return false
}

// OrgProviderEnv returns an empty typed provider environment.
func (noopProviderRuntime) OrgProviderEnv(pluginID string) ProviderEnv {
	return ProviderEnv{PluginID: pluginID}
}

// Available reports whether an active organization provider is available.
func (s *serviceImpl) Available(ctx context.Context) bool {
	if s == nil {
		return false
	}
	return defaultManager.StatusWithProvider(ctx, CapabilityOrgV1, s.runtime, s.providerEnv).Available
}

// Status returns the current organization capability activation state.
func (s *serviceImpl) Status(ctx context.Context) capmodel.CapabilityStatus {
	if s == nil {
		return convertCapabilityStatus(defaultManager.Status(ctx, CapabilityOrgV1, nil))
	}
	return convertCapabilityStatus(defaultManager.StatusWithProvider(ctx, CapabilityOrgV1, s.runtime, s.providerEnv))
}

// currentProvider returns the currently registered organization-capability provider.
func (s *serviceImpl) currentProvider(ctx context.Context) (Provider, error) {
	if s == nil {
		return nil, nil
	}
	provider, err := defaultManager.ActiveProviderWithError(ctx, CapabilityOrgV1, s.runtime, s.providerEnv)
	if err != nil || provider == nil {
		return nil, err
	}
	typedProvider, ok := provider.(Provider)
	if !ok {
		return nil, nil
	}
	return typedProvider, nil
}

// providerEnv builds lazy construction inputs for one organization provider.
func (s *serviceImpl) providerEnv(_ context.Context, pluginID string) ProviderEnv {
	env := ProviderEnv{PluginID: pluginID}
	if s != nil && s.runtime != nil {
		env = s.runtime.OrgProviderEnv(pluginID)
	}
	if env.PluginID == "" {
		env.PluginID = pluginID
	}
	return env
}

// ListUserDeptAssignments returns user -> department projections for the provided users.
func (s *serviceImpl) ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*UserDeptAssignment, error) {
	assignments := make(map[int]*UserDeptAssignment)
	if len(userIDs) == 0 {
		return assignments, nil
	}

	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return assignments, nil
	}
	return provider.ListUserDeptAssignments(ctx, userIDs)
}

// GetUserDeptInfo returns one user's department projection.
func (s *serviceImpl) GetUserDeptInfo(ctx context.Context, userID int) (int, string, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return 0, "", err
	}
	if provider == nil {
		return 0, "", nil
	}
	return provider.GetUserDeptInfo(ctx, userID)
}

// GetUserDeptName returns one user's department name for online-session projection.
func (s *serviceImpl) GetUserDeptName(ctx context.Context, userID int) (string, error) {
	_, deptName, err := s.GetUserDeptInfo(ctx, userID)
	return deptName, err
}

// GetUserDeptIDs returns one user's department identifier list.
func (s *serviceImpl) GetUserDeptIDs(ctx context.Context, userID int) ([]int, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return []int{}, nil
	}
	return provider.GetUserDeptIDs(ctx, userID)
}

// ApplyUserDeptScope injects a database-side department-scope constraint for
// rows owned by the supplied user ID column.
func (s *serviceImpl) ApplyUserDeptScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	currentUserID int,
) (*gdb.Model, bool, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, false, err
	}
	if provider == nil {
		return model, true, nil
	}
	return provider.ApplyUserDeptScope(ctx, model, userIDColumn, currentUserID)
}

// BuildUserDeptScopeExists builds the database-side department-scope EXISTS
// subquery for callers that need to compose it with additional OR branches.
func (s *serviceImpl) BuildUserDeptScopeExists(
	ctx context.Context,
	userIDColumn string,
	currentUserID int,
) (*gdb.Model, bool, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, false, err
	}
	if provider == nil {
		return nil, true, nil
	}
	return provider.BuildUserDeptScopeExists(ctx, userIDColumn, currentUserID)
}

// ApplyUserDeptFilter constrains user rows to a requested department subtree without materializing user IDs.
func (s *serviceImpl) ApplyUserDeptFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	deptID int,
) (*gdb.Model, bool, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, false, err
	}
	if provider == nil || model == nil {
		return model, false, nil
	}
	return provider.ApplyUserDeptFilter(ctx, model, userIDColumn, deptID)
}

// ApplyUserDeptUnassignedFilter constrains user rows to records without department assignments.
func (s *serviceImpl) ApplyUserDeptUnassignedFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
) (*gdb.Model, bool, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, false, err
	}
	if provider == nil || model == nil {
		return model, false, nil
	}
	return provider.ApplyUserDeptUnassignedFilter(ctx, model, userIDColumn)
}

// GetUserPostIDs returns one user's post association list.
func (s *serviceImpl) GetUserPostIDs(ctx context.Context, userID int) ([]int, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return []int{}, nil
	}
	return provider.GetUserPostIDs(ctx, userID)
}

// ReplaceUserAssignments rewrites one user's department and post associations.
func (s *serviceImpl) ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return nil
	}
	return provider.ReplaceUserAssignments(ctx, userID, deptID, postIDs)
}

// CleanupUserAssignments deletes one user's optional organization associations.
func (s *serviceImpl) CleanupUserAssignments(ctx context.Context, userID int) error {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return nil
	}
	return provider.CleanupUserAssignments(ctx, userID)
}

// UserDeptTree returns the optional department tree used by host user management.
func (s *serviceImpl) UserDeptTree(ctx context.Context) ([]*DeptTreeNode, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return []*DeptTreeNode{}, nil
	}
	return provider.UserDeptTree(ctx)
}

// ListPostOptions returns selectable post options for one department subtree.
func (s *serviceImpl) ListPostOptions(ctx context.Context, deptID *int) ([]*PostOption, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return []*PostOption{}, nil
	}
	return provider.ListPostOptions(ctx, deptID)
}
