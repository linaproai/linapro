// Package orgcap implements the optional host-side organization capability seam
// used by user-management and auth flows. The host keeps only the stable
// interface and delegates real organization behavior to one registered plugin
// provider when the org-center plugin is enabled.
package orgcap

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	pkgorgcap "lina-core/pkg/orgcap"
)

// UserDeptAssignment describes one optional department projection for a user.
type UserDeptAssignment = pkgorgcap.UserDeptAssignment

// DeptTreeNode is the host-owned department tree projection exposed through
// the orgcap capability seam.
type DeptTreeNode = pkgorgcap.DeptTreeNode

// PostOption describes one selectable post projection exposed through the
// organization capability seam for host-owned user-management flows.
type PostOption = pkgorgcap.PostOption

// PluginEnablementReader defines the narrow plugin state capability required by
// orgcap to detect whether the org-center plugin is currently enabled.
type PluginEnablementReader interface {
	// IsEnabled returns whether the given plugin ID is currently enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
}

// Service defines the optional organization capability consumed by host core services.
type Service interface {
	// Enabled reports whether organization capability is currently installed and enabled.
	Enabled(ctx context.Context) bool
	// ListUserDeptAssignments returns user -> department projections for the provided users.
	ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*UserDeptAssignment, error)
	// GetUserIDsByDept returns user IDs associated with the given department subtree.
	GetUserIDsByDept(ctx context.Context, deptID int) ([]int, error)
	// GetAllAssignedUserIDs returns all user IDs that currently hold department assignments.
	GetAllAssignedUserIDs(ctx context.Context) ([]int, error)
	// GetUserDeptInfo returns one user's department projection.
	GetUserDeptInfo(ctx context.Context, userID int) (int, string, error)
	// GetUserDeptName returns one user's department name for online-session projection.
	GetUserDeptName(ctx context.Context, userID int) (string, error)
	// GetUserDeptIDs returns one user's department identifier list.
	GetUserDeptIDs(ctx context.Context, userID int) ([]int, error)
	// ApplyUserDeptScope injects a database-side department-scope constraint for
	// rows owned by the supplied user ID column.
	ApplyUserDeptScope(ctx context.Context, model *gdb.Model, userIDColumn string, currentUserID int) (*gdb.Model, bool, error)
	// BuildUserDeptScopeExists builds the database-side department-scope EXISTS
	// subquery for callers that need to compose it with additional OR branches.
	BuildUserDeptScopeExists(ctx context.Context, userIDColumn string, currentUserID int) (*gdb.Model, bool, error)
	// GetUserPostIDs returns one user's post association list.
	GetUserPostIDs(ctx context.Context, userID int) ([]int, error)
	// ReplaceUserAssignments rewrites one user's department and post associations.
	ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error
	// CleanupUserAssignments deletes one user's optional organization associations.
	CleanupUserAssignments(ctx context.Context, userID int) error
	// UserDeptTree returns the optional department tree used by host user management.
	UserDeptTree(ctx context.Context) ([]*DeptTreeNode, error)
	// ListPostOptions returns selectable post options for one department subtree.
	ListPostOptions(ctx context.Context, deptID *int) ([]*PostOption, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	enablementReader PluginEnablementReader
}

// New creates and returns a new optional organization capability service.
// Pass a non-nil enablementReader to couple orgcap to plugin enablement state;
// pass nil to use the default reader that treats the capability as disabled.
func New(enablementReader PluginEnablementReader) Service {
	if enablementReader == nil {
		enablementReader = noopPluginEnablementReader{}
	}
	return &serviceImpl{
		enablementReader: enablementReader,
	}
}

// Enabled reports whether organization capability is currently installed and enabled.
func (s *serviceImpl) Enabled(ctx context.Context) bool {
	if s == nil || s.enablementReader == nil {
		return false
	}
	if !s.enablementReader.IsEnabled(ctx, pkgorgcap.ProviderPluginID) {
		return false
	}
	return pkgorgcap.HasProvider()
}

// noopPluginEnablementReader reports all plugins as disabled when orgcap is
// constructed without an explicit enablement reader.
type noopPluginEnablementReader struct{}

// IsEnabled always returns false.
func (noopPluginEnablementReader) IsEnabled(_ context.Context, _ string) bool {
	return false
}

// currentProvider returns the currently registered organization-capability provider.
func (s *serviceImpl) currentProvider(ctx context.Context) pkgorgcap.Provider {
	if !s.Enabled(ctx) {
		return nil
	}
	return pkgorgcap.CurrentProvider()
}

// ListUserDeptAssignments returns user -> department projections for the provided users.
func (s *serviceImpl) ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*UserDeptAssignment, error) {
	assignments := make(map[int]*UserDeptAssignment)
	if len(userIDs) == 0 {
		return assignments, nil
	}

	provider := s.currentProvider(ctx)
	if provider == nil {
		return assignments, nil
	}
	return provider.ListUserDeptAssignments(ctx, userIDs)
}

// GetUserIDsByDept returns user IDs associated with the given department subtree.
func (s *serviceImpl) GetUserIDsByDept(ctx context.Context, deptID int) ([]int, error) {
	provider := s.currentProvider(ctx)
	if provider == nil {
		return []int{}, nil
	}
	return provider.GetUserIDsByDept(ctx, deptID)
}

// GetAllAssignedUserIDs returns all user IDs that currently hold department assignments.
func (s *serviceImpl) GetAllAssignedUserIDs(ctx context.Context) ([]int, error) {
	provider := s.currentProvider(ctx)
	if provider == nil {
		return []int{}, nil
	}
	return provider.GetAllAssignedUserIDs(ctx)
}

// GetUserDeptInfo returns one user's department projection.
func (s *serviceImpl) GetUserDeptInfo(ctx context.Context, userID int) (int, string, error) {
	provider := s.currentProvider(ctx)
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
	provider := s.currentProvider(ctx)
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
	provider := s.currentProvider(ctx)
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
	provider := s.currentProvider(ctx)
	if provider == nil {
		return nil, true, nil
	}
	return provider.BuildUserDeptScopeExists(ctx, userIDColumn, currentUserID)
}

// GetUserPostIDs returns one user's post association list.
func (s *serviceImpl) GetUserPostIDs(ctx context.Context, userID int) ([]int, error) {
	provider := s.currentProvider(ctx)
	if provider == nil {
		return []int{}, nil
	}
	return provider.GetUserPostIDs(ctx, userID)
}

// ReplaceUserAssignments rewrites one user's department and post associations.
func (s *serviceImpl) ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error {
	provider := s.currentProvider(ctx)
	if provider == nil {
		return nil
	}
	return provider.ReplaceUserAssignments(ctx, userID, deptID, postIDs)
}

// CleanupUserAssignments deletes one user's optional organization associations.
func (s *serviceImpl) CleanupUserAssignments(ctx context.Context, userID int) error {
	provider := s.currentProvider(ctx)
	if provider == nil {
		return nil
	}
	return provider.CleanupUserAssignments(ctx, userID)
}

// UserDeptTree returns the optional department tree used by host user management.
func (s *serviceImpl) UserDeptTree(ctx context.Context) ([]*DeptTreeNode, error) {
	provider := s.currentProvider(ctx)
	if provider == nil {
		return []*DeptTreeNode{}, nil
	}
	return provider.UserDeptTree(ctx)
}

// ListPostOptions returns selectable post options for one department subtree.
func (s *serviceImpl) ListPostOptions(ctx context.Context, deptID *int) ([]*PostOption, error) {
	provider := s.currentProvider(ctx)
	if provider == nil {
		return []*PostOption{}, nil
	}
	return provider.ListPostOptions(ctx, deptID)
}
