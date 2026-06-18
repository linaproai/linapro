// Package orgcap owns the stable organization capability contract exposed
// through capability. Provider SPI, database scope helpers, and host workspace
// projection seams live in orgspi so ordinary consumers see only DTO contracts.
package orgcap

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
)

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

// UserDeptAssignment describes one optional department projection for a user.
type UserDeptAssignment struct {
	// DeptID is the associated department identifier.
	DeptID int
	// DeptName is the associated department display name.
	DeptName string
}

// DeptTreeNode is one host-facing department tree node projection.
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

// PostOption describes one selectable post projection exposed to host flows.
type PostOption struct {
	// PostID is the selectable post identifier.
	PostID int
	// PostName is the selectable post display name.
	PostName string
}

// UserOrgProfile describes one user's stable organization projection.
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

// DeptProjection describes one stable department candidate projection.
type DeptProjection struct {
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

// DeptTreeResult contains one bounded department tree projection.
type DeptTreeResult struct {
	// Items contains root department nodes.
	Items []*DeptTreeNode `json:"items"`
	// Total is the number of nodes before any max-node truncation.
	Total int `json:"total"`
	// Truncated reports whether the node list was truncated by MaxNodes.
	Truncated bool `json:"truncated"`
}

// DeptSearchInput describes bounded department candidate search.
type DeptSearchInput struct {
	// Keyword matches stable department name or code fields.
	Keyword string `json:"keyword,omitempty"`
	// Status optionally filters by provider-owned status.
	Status *int `json:"status,omitempty"`
	// Page constrains page number, page size and optional limit.
	Page capmodel.PageRequest `json:"page"`
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

// Service defines the optional organization capability consumed by host core
// services and plugins without depending on a concrete provider implementation.
//
// Service 定义宿主核心服务和普通插件可消费的只读组织能力，适用于读取用户部门、岗位等稳定组织投影，并在组织插件缺失时获得安全降级结果。
type Service interface {
	// Available reports whether an active organization provider is available.
	//
	// Available 判断当前是否存在可用组织能力提供方，适用于调用方决定展示、降级或跳过组织相关逻辑。
	Available(ctx context.Context) bool
	// Status returns the current organization capability activation state.
	//
	// Status 返回组织能力激活状态，适用于诊断、治理检查和插件能力状态展示。
	Status(ctx context.Context) capmodel.CapabilityStatus
	// ListUserDeptAssignments returns user-to-department projections for the provided users.
	//
	// ListUserDeptAssignments 批量返回用户部门归属投影，适用于列表、详情批量和导出等需要集合化装配部门信息的场景。
	ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*UserDeptAssignment, error)
	// BatchGetUserOrgProfiles returns stable organization profiles for visible users.
	//
	// BatchGetUserOrgProfiles 批量返回用户组织档案，适用于插件列表、详情批量和导出装配部门与岗位投影；provider 缺失时返回空档案。
	BatchGetUserOrgProfiles(ctx context.Context, userIDs []int) (*capmodel.BatchResult[*UserOrgProfile, int], error)
	// GetUserDeptInfo returns one user's department projection.
	//
	// GetUserDeptInfo 返回单个用户的部门标识和名称，适用于详情读取、会话补充和低频单用户查询场景。
	GetUserDeptInfo(ctx context.Context, userID int) (int, string, error)
	// GetUserDeptName returns one user's department name for online-session projection.
	//
	// GetUserDeptName 返回单个用户的部门名称，适用于在线会话、审计展示和只需要名称的轻量投影场景。
	GetUserDeptName(ctx context.Context, userID int) (string, error)
	// GetUserDeptIDs returns one user's department identifier list.
	//
	// GetUserDeptIDs 返回单个用户所属部门标识集合，适用于权限判定和组织范围计算场景。
	GetUserDeptIDs(ctx context.Context, userID int) ([]int, error)
	// GetUserPostIDs returns one user's post association list.
	//
	// GetUserPostIDs 返回单个用户关联岗位标识集合，适用于用户详情、编辑回显和组织关系读取场景。
	GetUserPostIDs(ctx context.Context, userID int) ([]int, error)
	// ListDeptTree returns a bounded department tree projection for ordinary plugins.
	//
	// ListDeptTree 返回有节点上限的部门树投影，适用于跨插件组织候选展示；provider 缺失时返回空树。
	ListDeptTree(ctx context.Context, input DeptTreeInput) (*DeptTreeResult, error)
	// SearchDepartments returns bounded department candidates.
	//
	// SearchDepartments 返回分页部门候选投影，适用于插件表单、筛选和关系选择；provider 缺失时返回空页。
	SearchDepartments(ctx context.Context, input DeptSearchInput) (*capmodel.PageResult[*DeptProjection], error)
	// ListPostOptionsPage returns bounded post candidates.
	//
	// ListPostOptionsPage 返回分页岗位候选投影，适用于插件表单、筛选和关系选择；provider 缺失时返回空页。
	ListPostOptionsPage(ctx context.Context, input PostOptionsInput) (*capmodel.PageResult[*PostOption], error)
	// EnsureDepartmentsVisible verifies every department identifier is visible to the caller.
	//
	// EnsureDepartmentsVisible 校验部门引用在当前租户和组织 provider 边界内可见，任一目标不可见时整体拒绝。
	EnsureDepartmentsVisible(ctx context.Context, deptIDs []int) error
	// EnsurePostsVisible verifies every post identifier is visible to the caller.
	//
	// EnsurePostsVisible 校验岗位引用在当前租户和组织 provider 边界内可见，任一目标不可见时整体拒绝。
	EnsurePostsVisible(ctx context.Context, postIDs []int) error
}
