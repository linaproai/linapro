// Package user implements user management, profile maintenance, import/export,
// and related authorization helpers for the Lina core host service.
package user

import (
	"context"
	"io"
	"sync"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/orgcap"
	"lina-core/internal/service/role"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/internal/service/user/accountpolicy"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/gdbutil"
	pkgtenantcap "lina-core/pkg/tenantcap"
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

var instance Service
var once sync.Once

// Instance returns the singleton user service instance.
// It initializes the instance exactly once, using default dependencies.
func Instance() Service {
	once.Do(func() {
		instance = New(nil)
	})
	return instance
}

// New creates and returns a new Service instance.
// Deprecated: Use Instance() for singleton access.
// The orgCapSvc parameter is optional; pass nil to use the default
// organization capability service; pass nil to use the default instance.
func New(orgCapSvc orgcap.Service, tenantEnablementReaders ...tenantcapsvc.PluginEnablementReader) Service {
	_ = tenantEnablementReaders // Keep for API compatibility
	if orgCapSvc == nil {
		orgCapSvc = orgcap.Instance()
	}
	svc := &serviceImpl{authSvc: auth.Instance(),
		bizCtxSvc: bizctx.Instance(),
		i18nSvc:   i18nsvc.Instance(),
		orgCapSvc: orgCapSvc,
		roleSvc:   role.Instance(),
		tenantSvc: tenantcapsvc.Instance(),
	}
	svc.scopeSvc = datascope.New(datascope.Dependencies{
		BizCtxSvc: svc.bizCtxSvc,
		RoleSvc:   svc.roleSvc,
		OrgCapSvc: svc.orgCapSvc,
	})
	return svc
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

// List queries user list with pagination and filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols       = dao.SysUser.Columns()
		m          = dao.SysUser.Ctx(ctx)
		orgEnabled = s.orgCapSvc.Enabled(ctx)
	)
	m, _, err := s.tenantSvc.ApplyUserTenantScope(ctx, m, qualifiedSysUserIDColumn())
	if err != nil {
		return nil, err
	}

	// Apply filters
	if in.Username != "" {
		m = m.WhereLike(cols.Username, "%"+in.Username+"%")
	}
	if in.Nickname != "" {
		m = m.WhereLike(cols.Nickname, "%"+in.Nickname+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}
	if in.Phone != "" {
		m = m.WhereLike(cols.Phone, "%"+in.Phone+"%")
	}
	if in.Sex != nil {
		m = m.Where(cols.Sex, *in.Sex)
	}
	if s.multiTenantEnabled(ctx) && in.TenantId != nil {
		if err = s.ensureListTenantFilterAllowed(ctx, *in.TenantId); err != nil {
			return nil, err
		}
		m, _, err = s.tenantSvc.ApplyUserTenantFilter(
			ctx,
			m,
			qualifiedSysUserIDColumn(),
			tenantcapsvc.TenantID(*in.TenantId),
		)
		if err != nil {
			return nil, err
		}
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.CreatedAt, in.BeginTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(cols.CreatedAt, in.EndTime)
	}

	// Filter by dept via association table
	if orgEnabled && in.DeptId != nil {
		if *in.DeptId == 0 {
			// Unassigned means users without any organization-assignment projection.
			assignedUserIds, err := s.GetAllAssignedUserIds(ctx)
			if err != nil {
				return nil, err
			}
			if len(assignedUserIds) > 0 {
				m = m.WhereNotIn(cols.Id, assignedUserIds)
			}
		} else {
			userIds, err := s.GetUserIdsByDeptId(ctx, *in.DeptId)
			if err != nil {
				return nil, err
			}
			if len(userIds) == 0 {
				return &ListOutput{List: []*ListOutputItem{}, Total: 0}, nil
			}
			m = m.WhereIn(cols.Id, userIds)
		}
	}

	m, scopeEmpty, err := s.applyUserDataScope(ctx, m)
	if err != nil {
		return nil, err
	}
	if scopeEmpty {
		return &ListOutput{List: []*ListOutputItem{}, Total: 0}, nil
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Normalize the requested sort field and direction before applying the
	// shared helper so business code never hand-builds ORDER BY fragments.
	var (
		allowedSortFields = map[string]string{
			"id":         cols.Id,
			"username":   cols.Username,
			"nickname":   cols.Nickname,
			"phone":      cols.Phone,
			"email":      cols.Email,
			"status":     cols.Status,
			"created_at": cols.CreatedAt,
			"createdAt":  cols.CreatedAt,
		}
		sortField     = cols.Id
		sortDirection = gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)
	)
	if f, ok := allowedSortFields[in.OrderBy]; ok {
		sortField = f
	}

	// Query with pagination, exclude password field
	var list []*entity.SysUser
	err = gdbutil.ApplyModelOrder(
		m.FieldsEx(cols.Password).Page(in.PageNum, in.PageSize),
		sortField,
		sortDirection,
	).Scan(&list)
	if err != nil {
		return nil, err
	}

	// Batch query dept info to avoid N+1 problem
	items := make([]*ListOutputItem, 0, len(list))
	if len(list) == 0 {
		return &ListOutput{List: items, Total: total}, nil
	}

	// Collect all user IDs
	userIds := make([]int, 0, len(list))
	for _, u := range list {
		userIds = append(userIds, u.Id)
	}

	userDeptMap := make(map[int]int)
	deptNameMap := make(map[int]string)
	if orgEnabled {
		assignments, assignmentErr := s.orgCapSvc.ListUserDeptAssignments(ctx, userIds)
		if assignmentErr != nil {
			return nil, assignmentErr
		}
		for userID, assignment := range assignments {
			if assignment == nil {
				continue
			}
			userDeptMap[userID] = assignment.DeptID
			deptNameMap[assignment.DeptID] = assignment.DeptName
		}
	}

	userTenantMap := make(map[int]*pkgtenantcap.UserTenantProjection)
	if s.multiTenantEnabled(ctx) {
		userTenantMap, err = s.tenantSvc.ListUserTenantProjections(ctx, userIds)
		if err != nil {
			return nil, err
		}
	}

	// Build user-role associations
	urCols := dao.SysUserRole.Columns()
	var userRoles []*entity.SysUserRole
	userRoleModel := dao.SysUserRole.Ctx(ctx).WhereIn(urCols.UserId, userIds)
	userRoleModel = datascope.ApplyTenantScope(ctx, userRoleModel, datascope.TenantColumn)
	err = userRoleModel.Scan(&userRoles)
	if err != nil {
		return nil, err
	}

	// Build userId -> roleIds map
	userRoleMap := make(map[int][]int)
	roleIdsSet := make(map[int]bool)
	for _, ur := range userRoles {
		userRoleMap[ur.UserId] = append(userRoleMap[ur.UserId], ur.RoleId)
		roleIdsSet[ur.RoleId] = true
	}

	// Get all unique role IDs
	allRoleIds := make([]int, 0, len(roleIdsSet))
	for roleId := range roleIdsSet {
		allRoleIds = append(allRoleIds, roleId)
	}

	// Batch query role info
	roleCols := dao.SysRole.Columns()
	var roles []*entity.SysRole
	if len(allRoleIds) > 0 {
		roleModel := dao.SysRole.Ctx(ctx).WhereIn(roleCols.Id, allRoleIds)
		roleModel = datascope.ApplyTenantScope(ctx, roleModel, datascope.TenantColumn)
		err = roleModel.Scan(&roles)
		if err != nil {
			return nil, err
		}
	}

	// Build roleId -> roleName map
	roleNameMap := make(map[int]string)
	for _, r := range roles {
		roleNameMap[r.Id] = s.roleSvc.DisplayName(ctx, r)
	}

	// Build output with dept and role info
	for _, u := range list {
		item := &ListOutputItem{SysUser: u}
		if deptId, ok := userDeptMap[u.Id]; ok {
			item.DeptId = deptId
			item.DeptName = deptNameMap[deptId]
		}
		// Get role info
		if roleIds, ok := userRoleMap[u.Id]; ok {
			item.RoleIds = roleIds
			for _, roleId := range roleIds {
				if name, exists := roleNameMap[roleId]; exists {
					item.RoleNames = append(item.RoleNames, name)
				}
			}
		} else {
			item.RoleIds = []int{}
			item.RoleNames = []string{}
		}
		if tenants, ok := userTenantMap[u.Id]; ok {
			for _, tenantID := range tenants.TenantIDs {
				item.TenantIds = append(item.TenantIds, int(tenantID))
			}
			item.TenantNames = tenants.TenantNames
		}
		items = append(items, item)
	}

	return &ListOutput{
		List:  items,
		Total: total,
	}, nil
}

// GetUserIdsByDeptId returns user IDs associated with a dept and all its descendants.
func (s *serviceImpl) GetUserIdsByDeptId(ctx context.Context, deptId int) ([]int, error) {
	return s.orgCapSvc.GetUserIDsByDept(ctx, deptId)
}

// GetAllAssignedUserIds returns all user IDs that have a dept association.
func (s *serviceImpl) GetAllAssignedUserIds(ctx context.Context) ([]int, error) {
	return s.orgCapSvc.GetAllAssignedUserIDs(ctx)
}

// GetUserDeptInfo returns the dept ID and name for a user.
func (s *serviceImpl) GetUserDeptInfo(ctx context.Context, userId int) (int, string, error) {
	return s.orgCapSvc.GetUserDeptInfo(ctx, userId)
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

// Create creates a new user with transaction support.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int, error) {
	// Check username uniqueness
	count, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Username: in.Username}).
		Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, bizerr.NewCode(CodeUserUsernameExists)
	}

	// Hash password
	hash, err := s.authSvc.HashPassword(in.Password)
	if err != nil {
		return 0, err
	}

	// Default nickname to username if empty
	nickname := in.Nickname
	if nickname == "" {
		nickname = in.Username
	}

	var userId int

	tenantPlan, err := s.resolveCreateTenantMemberships(ctx, in.TenantIds)
	if err != nil {
		return 0, err
	}
	primaryTenantID := int(tenantPlan.PrimaryTenant)

	// Use transaction to ensure atomicity
	err = dao.SysUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Insert user (GoFrame auto-fills created_at and updated_at)
		id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
			Username: in.Username,
			Password: hash,
			Nickname: nickname,
			Email:    in.Email,
			Phone:    in.Phone,
			Sex:      in.Sex,
			Status:   in.Status,
			Remark:   in.Remark,
			TenantId: primaryTenantID,
		}).InsertAndGetId()
		if err != nil {
			return err
		}

		userId = int(id)
		if tenantPlan.ShouldReplace {
			if err = s.tenantSvc.ReplaceUserTenantAssignments(ctx, userId, tenantPlan); err != nil {
				return err
			}
		}

		if err = s.orgCapSvc.ReplaceUserAssignments(ctx, userId, in.DeptId, in.PostIds); err != nil {
			return err
		}

		if err = s.replaceUserRoleAssignments(ctx, userId, in.RoleIds); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return userId, nil
}

// GetById retrieves user by ID.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*entity.SysUser, error) {
	if err := s.ensureUserVisible(ctx, id); err != nil {
		return nil, err
	}
	return s.getById(ctx, id)
}

// getById retrieves one user row without applying role data-scope. It is used
// by self-service paths after they have already resolved the target as the
// current authenticated user.
func (s *serviceImpl) getById(ctx context.Context, id int) (*entity.SysUser, error) {
	var user *entity.SysUser
	cols := dao.SysUser.Columns()
	err := dao.SysUser.Ctx(ctx).
		FieldsEx(cols.Password).
		Where(do.SysUser{Id: id}).
		Scan(&user)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, bizerr.NewCode(CodeUserNotFound)
	}
	return user, nil
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

// Update updates user information with transaction support.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	// Cannot edit self via admin panel
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx != nil && bizCtx.UserId == in.Id {
		return bizerr.NewCode(CodeUserCurrentEditDenied)
	}

	// Check user exists
	if _, err := s.GetById(ctx, in.Id); err != nil {
		return err
	}

	data := do.SysUser{}
	if in.Username != nil {
		data.Username = *in.Username
	}
	if in.Password != nil && *in.Password != "" {
		hash, err := s.authSvc.HashPassword(*in.Password)
		if err != nil {
			return err
		}
		data.Password = hash
	}
	if in.Nickname != nil {
		data.Nickname = *in.Nickname
	}
	if in.Email != nil {
		data.Email = *in.Email
	}
	if in.Phone != nil {
		data.Phone = *in.Phone
	}
	if in.Sex != nil {
		data.Sex = *in.Sex
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}
	tenantPlan, err := s.resolveUpdateTenantMemberships(ctx, in.TenantIds)
	if err != nil {
		return err
	}
	if tenantPlan.ShouldReplace {
		data.TenantId = int(tenantPlan.PrimaryTenant)
	}

	// Use transaction to ensure atomicity
	err = dao.SysUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Update user
		_, err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: in.Id}).Data(data).Update()
		if err != nil {
			return err
		}

		if tenantPlan.ShouldReplace {
			if err = s.tenantSvc.ReplaceUserTenantAssignments(ctx, in.Id, tenantPlan); err != nil {
				return err
			}
		}

		if in.DeptId != nil || in.PostIds != nil {
			if err = s.orgCapSvc.ReplaceUserAssignments(ctx, in.Id, in.DeptId, in.PostIds); err != nil {
				return err
			}
		}

		if in.RoleIds != nil {
			if err = s.replaceUserRoleAssignments(ctx, in.Id, in.RoleIds); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	s.roleSvc.NotifyAccessTopologyChanged(ctx)
	return nil
}

// Delete soft-deletes a user.
func (s *serviceImpl) Delete(ctx context.Context, id int) error {
	if err := s.runUserDeletionTransaction(ctx, []int{id}); err != nil {
		return err
	}
	s.roleSvc.NotifyAccessTopologyChanged(ctx)
	return nil
}

// BatchDelete soft-deletes multiple users atomically.
func (s *serviceImpl) BatchDelete(ctx context.Context, ids []int) error {
	normalizedIds := normalizeUserDeleteIDs(ids)
	if len(normalizedIds) == 0 {
		return bizerr.NewCode(CodeUserDeleteIdsRequired)
	}
	if err := s.runUserDeletionTransaction(ctx, normalizedIds); err != nil {
		return err
	}
	s.roleSvc.NotifyAccessTopologyChanged(ctx)
	return nil
}

// runUserDeletionTransaction validates and deletes users with all associations in one transaction.
func (s *serviceImpl) runUserDeletionTransaction(ctx context.Context, ids []int) error {
	return dao.SysUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		for _, id := range ids {
			if err := s.ensureUserDeleteAllowed(ctx, id); err != nil {
				return err
			}
		}
		for _, id := range ids {
			if err := s.deleteUserRecordAndAssociations(ctx, id); err != nil {
				return err
			}
		}
		return nil
	})
}

// ensureUserDeleteAllowed enforces built-in and current-user deletion protection.
func (s *serviceImpl) ensureUserDeleteAllowed(ctx context.Context, id int) error {
	user, err := s.GetById(ctx, id)
	if err != nil {
		return err
	}
	if accountpolicy.IsBuiltInAdminUsername(user.Username) {
		return bizerr.NewCode(CodeUserBuiltinAdminDeleteDenied)
	}

	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx != nil && bizCtx.UserId == id {
		return bizerr.NewCode(CodeUserCurrentDeleteDenied)
	}
	return nil
}

// deleteUserRecordAndAssociations soft-deletes one user and clears dependent associations.
func (s *serviceImpl) deleteUserRecordAndAssociations(ctx context.Context, id int) error {
	if _, err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: id}).Delete(); err != nil {
		return err
	}
	if err := s.orgCapSvc.CleanupUserAssignments(ctx, id); err != nil {
		return err
	}
	urCols := dao.SysUserRole.Columns()
	deleteModel := dao.SysUserRole.Ctx(ctx).Where(urCols.UserId, id)
	deleteModel = datascope.ApplyTenantScope(ctx, deleteModel, datascope.TenantColumn)
	_, err := deleteModel.Delete()
	return err
}

// normalizeUserDeleteIDs removes duplicate IDs while preserving request order.
func normalizeUserDeleteIDs(ids []int) []int {
	normalizedIds := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalizedIds = append(normalizedIds, id)
	}
	return normalizedIds
}

// UpdateStatus updates user status.
func (s *serviceImpl) UpdateStatus(ctx context.Context, id int, status Status) error {
	// Cannot disable self
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx != nil && bizCtx.UserId == id && status == StatusDisabled {
		return bizerr.NewCode(CodeUserCurrentDisableDenied)
	}

	if _, err := s.GetById(ctx, id); err != nil {
		return err
	}

	_, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: id}).
		Data(do.SysUser{
			Status: status,
		}).
		Update()
	if err != nil {
		return err
	}
	s.roleSvc.NotifyAccessTopologyChanged(ctx)
	return nil
}

// GetProfile retrieves current user profile.
func (s *serviceImpl) GetProfile(ctx context.Context) (*entity.SysUser, error) {
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil {
		return nil, bizerr.NewCode(CodeUserNotAuthenticated)
	}
	return s.getById(ctx, bizCtx.UserId)
}

// UpdateProfileInput defines input for UpdateProfile function.
type UpdateProfileInput struct {
	Nickname *string // Nickname
	Email    *string // Email
	Phone    *string // Phone number
	Sex      *int    // Gender: 0=Unknown 1=Male 2=Female
	Password *string // Password
}

// UpdateProfile updates current user profile.
func (s *serviceImpl) UpdateProfile(ctx context.Context, in UpdateProfileInput) error {
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil {
		return bizerr.NewCode(CodeUserNotAuthenticated)
	}

	data := do.SysUser{}
	if in.Nickname != nil {
		data.Nickname = *in.Nickname
	}
	if in.Email != nil {
		data.Email = *in.Email
	}
	if in.Phone != nil {
		data.Phone = *in.Phone
	}
	if in.Sex != nil {
		data.Sex = *in.Sex
	}
	if in.Password != nil && *in.Password != "" {
		hash, err := s.authSvc.HashPassword(*in.Password)
		if err != nil {
			return err
		}
		data.Password = hash
	}

	_, err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: bizCtx.UserId}).Data(data).Update()
	return err
}

// ResetPassword resets a user's password.
func (s *serviceImpl) ResetPassword(ctx context.Context, id int, password string) error {
	// Check user exists
	if _, err := s.GetById(ctx, id); err != nil {
		return err
	}

	// Hash password
	hash, err := s.authSvc.HashPassword(password)
	if err != nil {
		return err
	}

	_, err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: id}).
		Data(do.SysUser{
			Password: hash,
		}).
		Update()
	return err
}

// UpdateAvatar updates current user's avatar URL.
func (s *serviceImpl) UpdateAvatar(ctx context.Context, avatarUrl string) error {
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil {
		return bizerr.NewCode(CodeUserNotAuthenticated)
	}
	_, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: bizCtx.UserId}).
		Data(do.SysUser{
			Avatar: avatarUrl,
		}).
		Update()
	return err
}

// GetUserPostIds returns the post IDs associated with a user.
func (s *serviceImpl) GetUserPostIds(ctx context.Context, userId int) ([]int, error) {
	return s.orgCapSvc.GetUserPostIDs(ctx, userId)
}

// GetUserRoleIds returns the role IDs associated with a user.
func (s *serviceImpl) GetUserRoleIds(ctx context.Context, userId int) ([]int, error) {
	var userRoles []*entity.SysUserRole
	model := dao.SysUserRole.Ctx(ctx).Where(dao.SysUserRole.Columns().UserId, userId)
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	err := model.Scan(&userRoles)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(userRoles))
	for _, ur := range userRoles {
		ids = append(ids, ur.RoleId)
	}
	return ids, nil
}
