// Package orgcap owns the stable organization capability contract exposed
// through capability. Provider SPI, database scope helpers, and host workspace
// view seams live in orgspi so ordinary consumers see only DTO contracts.
package orgcap

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
)

// Service defines the optional organization capability consumed by host core
// services and plugins without depending on a concrete provider implementation.
type Service interface {
	// Available reports whether an active organization provider is available.
	Available(ctx context.Context) bool
	// Status returns the current organization capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// Department returns department resource operations.
	Department() DepartmentService
	// Post returns post resource operations.
	Post() PostService
	// Assignment returns user organization assignment operations.
	Assignment() AssignmentService
}

// DepartmentService defines the plugin-visible organization department resource.
type DepartmentService interface {
	// Get returns one visible department info record.
	Get(ctx context.Context, deptID int) (*DeptInfo, error)
	// BatchGet returns visible department info records and opaque missing IDs.
	BatchGet(ctx context.Context, deptIDs []int) (*capmodel.BatchResult[*DeptInfo, int], error)
	// List returns bounded department candidates.
	List(ctx context.Context, input DeptListInput) (*capmodel.PageResult[*DeptInfo], error)
	// ListTree returns a bounded department tree.
	ListTree(ctx context.Context, input DeptTreeInput) (*DeptTreeResult, error)
	// ListOptions returns bounded department options.
	ListOptions(ctx context.Context, input DeptOptionsInput) (*capmodel.PageResult[*DeptInfo], error)
	// EnsureVisible verifies department identifiers are visible to the caller.
	EnsureVisible(ctx context.Context, deptIDs []int) error
	// Create creates one department through the organization owner.
	Create(ctx context.Context, input DeptCreateInput) (int, error)
	// Update updates one department through the organization owner.
	Update(ctx context.Context, input DeptUpdateInput) error
	// Delete deletes one department through the organization owner.
	Delete(ctx context.Context, deptID int) error
}

// PostService defines the plugin-visible organization post resource.
type PostService interface {
	// Get returns one visible post info record.
	Get(ctx context.Context, postID int) (*PostInfo, error)
	// BatchGet returns visible post info records and opaque missing IDs.
	BatchGet(ctx context.Context, postIDs []int) (*capmodel.BatchResult[*PostInfo, int], error)
	// List returns bounded post candidates.
	List(ctx context.Context, input PostListInput) (*capmodel.PageResult[*PostInfo], error)
	// ListOptions returns bounded post options.
	ListOptions(ctx context.Context, input PostOptionsInput) (*capmodel.PageResult[*PostOption], error)
	// EnsureVisible verifies post identifiers are visible to the caller.
	EnsureVisible(ctx context.Context, postIDs []int) error
	// Create creates one post through the organization owner.
	Create(ctx context.Context, input PostCreateInput) (int, error)
	// Update updates one post through the organization owner.
	Update(ctx context.Context, input PostUpdateInput) error
	// Delete deletes one post through the organization owner.
	Delete(ctx context.Context, postID int) error
}

// AssignmentService defines plugin-visible user organization assignments.
type AssignmentService interface {
	// BatchGetUserProfiles returns stable organization profiles for visible users.
	BatchGetUserProfiles(ctx context.Context, userIDs []int) (*capmodel.BatchResult[*UserOrgProfile, int], error)
	// ListByUser returns one user's organization profile.
	ListByUser(ctx context.Context, userID int) (*UserOrgProfile, error)
	// BatchListByUsers returns user-to-department info records for the provided users.
	BatchListByUsers(ctx context.Context, userIDs []int) (map[int]*UserDeptAssignment, error)
	// GetUserDeptInfo returns one user's department information.
	GetUserDeptInfo(ctx context.Context, userID int) (int, string, error)
	// GetUserDeptIDs returns one user's department identifier list.
	GetUserDeptIDs(ctx context.Context, userID int) ([]int, error)
	// GetUserPostIDs returns one user's post association list.
	GetUserPostIDs(ctx context.Context, userID int) ([]int, error)
	// ReplaceByUser rewrites one user's department and post associations.
	ReplaceByUser(ctx context.Context, userID int, deptID *int, postIDs []int) error
	// CleanupByUser deletes one user's optional organization associations.
	CleanupByUser(ctx context.Context, userID int) error
}

const (
	// CapabilityOrgV1 identifies the versioned organization framework capability.
	CapabilityOrgV1 = "framework.org.v1"
	// ProviderPluginID is the official source-plugin identifier that provides organization capability.
	ProviderPluginID = "linapro-org-core"
)

const (
	// MaxUserOrgProfileBatchSize is the maximum user count accepted by batch profile reads.
	MaxUserOrgProfileBatchSize = 200
	// MaxDeptTreeNodes is the maximum department node count returned through the ordinary tree contract.
	MaxDeptTreeNodes = 500
	// MaxDeptSearchPageSize is the maximum department candidate page size.
	MaxDeptSearchPageSize = 200
	// MaxPostOptionsPageSize is the maximum post candidate page size.
	MaxPostOptionsPageSize = 200
	// MaxVisibilityCheckSize is the maximum department or post identifiers checked in one call.
	MaxVisibilityCheckSize = 200
)

// UserDeptAssignment describes one optional department assignment for a user.
type UserDeptAssignment struct {
	// DeptID is the associated department identifier.
	DeptID int
	// DeptName is the associated department display name.
	DeptName string
}

// DeptTreeNode is one host-facing department tree node.
type DeptTreeNode struct {
	// Id is the department identifier, or 0 for the synthetic unassigned node.
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

// PostOption describes one selectable post option exposed to host flows.
type PostOption struct {
	// PostID is the selectable post identifier.
	PostID int
	// PostName is the selectable post display name.
	PostName string
}

// UserOrgProfile describes one user's stable organization profile.
type UserOrgProfile struct {
	// UserID is the host user identifier this profile belongs to.
	UserID int `json:"userId"`
	// DeptID is the user's primary department identifier when assigned.
	DeptID int `json:"deptId"`
	// DeptName is the user's primary department display name when assigned.
	DeptName string `json:"deptName"`
	// PostIDs are visible post identifiers assigned to the user.
	PostIDs []int `json:"postIds"`
	// PostNames are visible post display names assigned to the user.
	PostNames []string `json:"postNames"`
}

// DeptInfo describes one stable department candidate.
type DeptInfo struct {
	// DeptID is the department identifier.
	DeptID int `json:"deptId"`
	// ParentID is the parent department identifier.
	ParentID int `json:"parentId"`
	// DeptName is the department display name.
	DeptName string `json:"deptName"`
	// DeptCode is the stable department code.
	DeptCode string `json:"deptCode"`
	// Status is the provider-owned department status.
	Status int `json:"status"`
}

// DeptTreeInput constrains ordinary department tree reads.
type DeptTreeInput struct {
	// MaxNodes caps the number of returned nodes; zero uses MaxDeptTreeNodes.
	MaxNodes int `json:"maxNodes,omitempty"`
}

// DeptTreeResult contains one bounded department tree.
type DeptTreeResult struct {
	// Items contains root department nodes.
	Items []*DeptTreeNode `json:"items"`
	// Total is the number of nodes before any max-node truncation.
	Total int `json:"total"`
	// Truncated reports whether the node list was truncated by MaxNodes.
	Truncated bool `json:"truncated"`
}

// DeptListInput describes bounded department candidate listing.
type DeptListInput struct {
	// Keyword matches stable department name or code fields.
	Keyword string `json:"keyword,omitempty"`
	// Status optionally filters by provider-owned status.
	Status *int `json:"status,omitempty"`
	// Page constrains page number, page size and optional limit.
	Page capmodel.PageRequest `json:"page"`
}

// DeptOptionsInput describes bounded department option listing.
type DeptOptionsInput struct {
	// Keyword matches stable department name or code fields.
	Keyword string `json:"keyword,omitempty"`
	// Status optionally filters by provider-owned status.
	Status *int `json:"status,omitempty"`
	// Page constrains page number, page size and optional limit.
	Page capmodel.PageRequest `json:"page"`
}

// DeptCreateInput carries plugin-visible department creation fields.
type DeptCreateInput struct {
	// ParentID is the parent department identifier; zero creates a root department.
	ParentID int `json:"parentId,omitempty"`
	// DeptName is the department display name.
	DeptName string `json:"deptName"`
	// DeptCode is the stable department code.
	DeptCode string `json:"deptCode,omitempty"`
	// OrderNum is the provider-owned display order.
	OrderNum int `json:"orderNum,omitempty"`
	// LeaderUserID optionally references the department leader user.
	LeaderUserID int `json:"leaderUserId,omitempty"`
	// Phone is the department contact phone.
	Phone string `json:"phone,omitempty"`
	// Email is the department contact email.
	Email string `json:"email,omitempty"`
	// Status is the provider-owned department status.
	Status int `json:"status,omitempty"`
	// Remark is an optional department note.
	Remark string `json:"remark,omitempty"`
}

// DeptUpdateInput carries plugin-visible department update fields.
type DeptUpdateInput struct {
	// DeptID is the department identifier to update.
	DeptID int `json:"deptId"`
	// ParentID optionally moves the department under another parent.
	ParentID *int `json:"parentId,omitempty"`
	// DeptName optionally updates the department display name.
	DeptName *string `json:"deptName,omitempty"`
	// DeptCode optionally updates the stable department code.
	DeptCode *string `json:"deptCode,omitempty"`
	// OrderNum optionally updates provider-owned display order.
	OrderNum *int `json:"orderNum,omitempty"`
	// LeaderUserID optionally updates the department leader user.
	LeaderUserID *int `json:"leaderUserId,omitempty"`
	// Phone optionally updates the department contact phone.
	Phone *string `json:"phone,omitempty"`
	// Email optionally updates the department contact email.
	Email *string `json:"email,omitempty"`
	// Status optionally updates provider-owned department status.
	Status *int `json:"status,omitempty"`
	// Remark optionally updates the department note.
	Remark *string `json:"remark,omitempty"`
}

// PostInfo describes one stable post candidate.
type PostInfo struct {
	// PostID is the post identifier.
	PostID int `json:"postId"`
	// DeptID is the department identifier that owns the post.
	DeptID int `json:"deptId"`
	// PostCode is the stable post code.
	PostCode string `json:"postCode"`
	// PostName is the post display name.
	PostName string `json:"postName"`
	// Sort is the provider-owned display order.
	Sort int `json:"sort"`
	// Status is the provider-owned post status.
	Status int `json:"status"`
	// Remark is an optional post note.
	Remark string `json:"remark,omitempty"`
}

// PostOptionsInput describes bounded post candidate reads.
type PostOptionsInput struct {
	// DeptID optionally restricts posts to one department subtree.
	DeptID *int `json:"deptId,omitempty"`
	// Keyword matches stable post name or code fields.
	Keyword string `json:"keyword,omitempty"`
	// Status optionally filters by provider-owned status.
	Status *int `json:"status,omitempty"`
	// Page constrains page number, page size and optional limit.
	Page capmodel.PageRequest `json:"page"`
}

// PostListInput describes bounded post candidate listing.
type PostListInput struct {
	// DeptID optionally restricts posts to one department subtree.
	DeptID *int `json:"deptId,omitempty"`
	// Keyword matches stable post name or code fields.
	Keyword string `json:"keyword,omitempty"`
	// Status optionally filters by provider-owned status.
	Status *int `json:"status,omitempty"`
	// Page constrains page number, page size and optional limit.
	Page capmodel.PageRequest `json:"page"`
}

// PostCreateInput carries plugin-visible post creation fields.
type PostCreateInput struct {
	// DeptID is the owning department identifier.
	DeptID int `json:"deptId"`
	// PostCode is the stable post code.
	PostCode string `json:"postCode"`
	// PostName is the post display name.
	PostName string `json:"postName"`
	// Sort is the provider-owned display order.
	Sort int `json:"sort,omitempty"`
	// Status is the provider-owned post status.
	Status int `json:"status,omitempty"`
	// Remark is an optional post note.
	Remark string `json:"remark,omitempty"`
}

// PostUpdateInput carries plugin-visible post update fields.
type PostUpdateInput struct {
	// PostID is the post identifier to update.
	PostID int `json:"postId"`
	// DeptID optionally changes the owning department.
	DeptID *int `json:"deptId,omitempty"`
	// PostCode optionally updates the stable post code.
	PostCode *string `json:"postCode,omitempty"`
	// PostName optionally updates the post display name.
	PostName *string `json:"postName,omitempty"`
	// Sort optionally updates provider-owned display order.
	Sort *int `json:"sort,omitempty"`
	// Status optionally updates provider-owned post status.
	Status *int `json:"status,omitempty"`
	// Remark optionally updates the post note.
	Remark *string `json:"remark,omitempty"`
}
