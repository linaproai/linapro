// Package orgcap implements the optional organization capability seam used by
// host user-management and auth flows so the host depends on one stable
// contract instead of directly coupling to the org-management plugin tables.
package orgcap

import (
	"context"

	"lina-core/internal/plugingovernance"
	deptsvc "lina-core/internal/service/dept"
	pluginsvc "lina-core/internal/service/plugin"
)

// UserDeptAssignment describes one optional department projection for a user.
type UserDeptAssignment struct {
	// DeptID is the associated department identifier.
	DeptID int
	// DeptName is the associated department display name.
	DeptName string
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
	// GetUserPostIDs returns one user's post association list.
	GetUserPostIDs(ctx context.Context, userID int) ([]int, error)
	// ReplaceUserAssignments rewrites one user's department and post associations.
	ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error
	// CleanupUserAssignments deletes one user's optional organization associations.
	CleanupUserAssignments(ctx context.Context, userID int) error
	// UserDeptTree returns the optional department tree used by host user management.
	UserDeptTree(ctx context.Context) ([]*deptsvc.TreeNode, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	deptSvc   deptsvc.Service
	pluginSvc pluginsvc.Service
}

// New creates and returns a new optional organization capability service.
func New() Service {
	return &serviceImpl{
		deptSvc:   deptsvc.New(),
		pluginSvc: pluginsvc.New(),
	}
}

// Enabled reports whether organization capability is currently installed and enabled.
func (s *serviceImpl) Enabled(ctx context.Context) bool {
	if s == nil || s.pluginSvc == nil {
		return false
	}
	return s.pluginSvc.IsEnabled(ctx, plugingovernance.OrgManagement)
}

// storageInstalled reports whether the organization plugin has already installed
// its optional storage tables into the host.
func (s *serviceImpl) storageInstalled(ctx context.Context) bool {
	if s == nil || s.pluginSvc == nil {
		return false
	}
	return s.pluginSvc.IsInstalled(ctx, plugingovernance.OrgManagement)
}
