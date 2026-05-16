// Package user implements user management, profile maintenance, import/export,
// and related authorization helpers for the Lina core host service.
package user

import (
	"context"
	"io"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/orgcap"
	"lina-core/internal/service/role"
	tenantcapsvc "lina-core/internal/service/tenantcap"
)

// Status represents user account status.
type Status int

// User status and default account constants.
const (
	// StatusNormal represents a normal user status.
	StatusNormal Status = 1

	// StatusDisabled represents a disabled user status.
	StatusDisabled Status = 0
)

// Service defines the user service contract.
type Service interface {
	// List queries user list with pagination and filters.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetUserIdsByDeptId returns user IDs associated with a dept and all its descendants.
	GetUserIdsByDeptId(ctx context.Context, deptId int) ([]int, error)
	// GetAllAssignedUserIds returns all user IDs that have a dept association.
	GetAllAssignedUserIds(ctx context.Context) ([]int, error)
	// GetUserDeptInfo returns the dept ID and name for a user.
	GetUserDeptInfo(ctx context.Context, userId int) (int, string, error)
	// Create creates a new user with transaction support.
	Create(ctx context.Context, in CreateInput) (int, error)
	// GetById retrieves user by ID.
	GetById(ctx context.Context, id int) (*entity.SysUser, error)
	// Update updates user information with transaction support.
	Update(ctx context.Context, in UpdateInput) error
	// Delete soft-deletes a user.
	Delete(ctx context.Context, id int) error
	// BatchDelete soft-deletes multiple users atomically.
	BatchDelete(ctx context.Context, ids []int) error
	// BatchUpdate updates selected users atomically.
	BatchUpdate(ctx context.Context, in BatchUpdateInput) error
	// UpdateStatus updates user status.
	UpdateStatus(ctx context.Context, id int, status Status) error
	// GetProfile retrieves current user profile.
	GetProfile(ctx context.Context) (*entity.SysUser, error)
	// UpdateProfile updates current user profile.
	UpdateProfile(ctx context.Context, in UpdateProfileInput) error
	// ResetPassword resets a user's password.
	ResetPassword(ctx context.Context, id int, password string) error
	// UpdateAvatar updates current user's avatar URL.
	UpdateAvatar(ctx context.Context, avatarUrl string) error
	// GetUserPostIds returns the post IDs associated with a user.
	GetUserPostIds(ctx context.Context, userId int) ([]int, error)
	// GetUserRoleIds returns the role IDs associated with a user.
	GetUserRoleIds(ctx context.Context, userId int) ([]int, error)
	// GetUserTenantMemberships returns the tenant IDs and names associated with a user.
	GetUserTenantMemberships(ctx context.Context, userId int) ([]int, []string, error)
	// Export generates an Excel file with user data based on IDs.
	Export(ctx context.Context, in ExportInput) (data []byte, err error)
	// Import reads an Excel file and creates users from it.
	Import(ctx context.Context, fileReader io.Reader) (result *ImportResult, err error)
	// GenerateImportTemplate creates an Excel template for user import.
	GenerateImportTemplate(ctx context.Context) (data []byte, err error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	authSvc   auth.Service
	bizCtxSvc bizctx.Service
	i18nSvc   userI18nTranslator
	orgCapSvc orgcap.Service
	roleSvc   role.Service // Role service
	scopeSvc  datascope.Service
	tenantSvc tenantcapsvc.Service
}

// New creates and returns a new user service from explicit runtime-owned dependencies.
func New(authSvc auth.Service, bizCtxSvc bizctx.Service, i18nSvc userI18nTranslator, orgCapSvc orgcap.Service, roleSvc role.Service, scopeSvc datascope.Service, tenantSvc tenantcapsvc.Service) Service {
	return &serviceImpl{
		authSvc:   authSvc,
		bizCtxSvc: bizCtxSvc,
		i18nSvc:   i18nSvc,
		orgCapSvc: orgCapSvc,
		roleSvc:   roleSvc,
		scopeSvc:  scopeSvc,
		tenantSvc: tenantSvc,
	}
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum        int    // Page number, starting from 1
	PageSize       int    // Items per page
	Username       string // Username, supports fuzzy search
	Nickname       string // Nickname, supports fuzzy search
	Status         *int   // Status: 1=Normal 0=Disabled
	Phone          string // Phone number, supports fuzzy search
	Sex            *int   // Gender: 0=Unknown 1=Male 2=Female
	DeptId         *int   // Department ID, 0 means unassigned
	TenantId       *int   // Tenant ID filter for platform context
	BeginTime      string // Creation time start
	EndTime        string // Creation time end
	OrderBy        string // Sort field
	OrderDirection string // Sort direction: asc/desc
}

// ListOutputItem defines a single item in list output with dept info.
type ListOutputItem struct {
	SysUser     *entity.SysUser // User entity
	DeptId      int             // Department ID
	DeptName    string          // Department name
	RoleIds     []int           // Role ID list
	RoleNames   []string        // Role name list
	TenantIds   []int           // Tenant ID list
	TenantNames []string        // Tenant name list
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*ListOutputItem // User list
	Total int               // Total count
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Username  string // Username
	Password  string // Password
	Nickname  string // Nickname
	Email     string // Email
	Phone     string // Phone number
	Sex       int    // Gender: 0=Unknown 1=Male 2=Female
	Status    Status // Status: StatusNormal=Normal StatusDisabled=Disabled
	Remark    string // Remark
	DeptId    *int   // Department ID
	PostIds   []int  // Post ID list
	RoleIds   []int  // Role ID list
	TenantIds []int  // Tenant ID list
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id        int     // User ID
	Username  *string // Username
	Password  *string // Password
	Nickname  *string // Nickname
	Email     *string // Email
	Phone     *string // Phone number
	Sex       *int    // Gender: 0=Unknown 1=Male 2=Female
	Status    *int    // Status: 1=Normal 0=Disabled
	Remark    *string // Remark
	DeptId    *int    // Department ID
	PostIds   []int   // Post ID list
	RoleIds   []int   // Role ID list
	TenantIds []int   // Tenant ID list
}

// UpdateProfileInput defines input for UpdateProfile function.
type UpdateProfileInput struct {
	Nickname *string // Nickname
	Email    *string // Email
	Phone    *string // Phone number
	Sex      *int    // Gender: 0=Unknown 1=Male 2=Female
	Password *string // Password
}
