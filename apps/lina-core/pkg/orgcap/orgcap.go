// Package orgcap publishes the stable organization-capability contract shared
// between the host and source plugins so plugin-owned organization data can be
// served through one explicit provider instead of leaking plugin tables into
// host internals.
package orgcap

import (
	"context"
	"sync"
)

// UserDeptAssignment describes one optional department projection for a user.
type UserDeptAssignment struct {
	// DeptID is the associated department identifier.
	DeptID int
	// DeptName is the associated department display name.
	DeptName string
}

// DeptTreeNode is one host-facing department tree node projection.
type DeptTreeNode struct {
	// Id is the department identifier, or 0 for the synthetic "unassigned" node.
	Id int `json:"id"`
	// Label is the display name of the department node.
	Label string `json:"label"`
	// LabelKey is an optional runtime i18n key for host-owned synthetic labels.
	LabelKey string `json:"labelKey,omitempty"`
	// UserCount is the number of users attached to this node.
	UserCount int `json:"userCount"`
	// Children lists nested department nodes under this entry.
	Children []*DeptTreeNode `json:"children"`
}

// PostOption describes one selectable post projection exposed to host flows.
type PostOption struct {
	// PostID is the selectable post identifier.
	PostID int
	// PostName is the selectable post display name.
	PostName string
}

// Provider defines the stable organization capability implemented by plugins.
type Provider interface {
	// ListUserDeptAssignments returns user -> department projections for the provided users.
	ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*UserDeptAssignment, error)
	// GetUserIDsByDept returns user IDs associated with the given department subtree.
	GetUserIDsByDept(ctx context.Context, deptID int) ([]int, error)
	// GetAllAssignedUserIDs returns all user IDs that currently hold department assignments.
	GetAllAssignedUserIDs(ctx context.Context) ([]int, error)
	// GetUserDeptInfo returns one user's department projection.
	GetUserDeptInfo(ctx context.Context, userID int) (int, string, error)
	// GetUserDeptIDs returns one user's department identifier list.
	GetUserDeptIDs(ctx context.Context, userID int) ([]int, error)
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

var (
	providerMu sync.RWMutex
	provider   Provider
)

// RegisterProvider publishes the current organization-capability provider.
func RegisterProvider(p Provider) {
	providerMu.Lock()
	defer providerMu.Unlock()

	provider = p
}

// CurrentProvider returns the registered organization-capability provider.
func CurrentProvider() Provider {
	providerMu.RLock()
	defer providerMu.RUnlock()

	return provider
}

// HasProvider reports whether one organization-capability provider is registered.
func HasProvider() bool {
	return CurrentProvider() != nil
}
