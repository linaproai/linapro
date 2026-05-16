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

// Service defines the optional organization capability consumed by host core
// services without hard-linking user, auth, or data-scope code to org-center.
type Service interface {
	// Enabled reports whether organization capability is installed, enabled, and
	// backed by a registered provider. Disabled mode must return empty fallback
	// data instead of failing host flows.
	Enabled(ctx context.Context) bool
	// ListUserDeptAssignments returns user-to-department projections for the
	// provided users; missing providers return an empty map.
	ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*UserDeptAssignment, error)
	// GetUserIDsByDept returns user IDs associated with the given department
	// subtree for list filtering and data-scope checks.
	GetUserIDsByDept(ctx context.Context, deptID int) ([]int, error)
	// GetAllAssignedUserIDs returns all user IDs that currently hold department
	// assignments, used to derive unassigned-user filters.
	GetAllAssignedUserIDs(ctx context.Context) ([]int, error)
	// GetUserDeptInfo returns one user's department projection. Disabled orgcap
	// returns zero ID and empty name.
	GetUserDeptInfo(ctx context.Context, userID int) (int, string, error)
	// GetUserDeptName returns one user's department name for online-session
	// projection without exposing department IDs to session callers.
	GetUserDeptName(ctx context.Context, userID int) (string, error)
	// GetUserDeptIDs returns one user's department identifier list for access
	// snapshot and department-scope resolution.
	GetUserDeptIDs(ctx context.Context, userID int) ([]int, error)
	// ApplyUserDeptScope injects a database-side department-scope constraint for
	// rows owned by the supplied user ID column and reports empty when no rows
	// can be visible.
	ApplyUserDeptScope(ctx context.Context, model *gdb.Model, userIDColumn string, currentUserID int) (*gdb.Model, bool, error)
	// BuildUserDeptScopeExists builds the database-side department-scope EXISTS
	// subquery for callers that need to compose it with additional OR branches.
	BuildUserDeptScopeExists(ctx context.Context, userIDColumn string, currentUserID int) (*gdb.Model, bool, error)
	// GetUserPostIDs returns one user's post association list; disabled orgcap
	// returns an empty list.
	GetUserPostIDs(ctx context.Context, userID int) ([]int, error)
	// ReplaceUserAssignments rewrites one user's department and post
	// associations through the provider; disabled orgcap is a no-op.
	ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error
	// CleanupUserAssignments deletes one user's optional organization
	// associations during user deletion; disabled orgcap is a no-op.
	CleanupUserAssignments(ctx context.Context, userID int) error
	// UserDeptTree returns the optional department tree used by host user
	// management; disabled orgcap returns an empty tree.
	UserDeptTree(ctx context.Context) ([]*DeptTreeNode, error)
	// ListPostOptions returns selectable post options for one department subtree
	// and follows provider-side visibility rules.
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

// noopPluginEnablementReader reports all plugins as disabled when orgcap is
// constructed without an explicit enablement reader.
type noopPluginEnablementReader struct{}
