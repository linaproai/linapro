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
}
