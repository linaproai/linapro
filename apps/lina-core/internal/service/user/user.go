// Package user implements user management, profile maintenance, import/export,
// and related authorization helpers for the Lina core host service.
package user

import (
	"context"
	"io"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/dept"
	"lina-core/internal/service/role"
	"lina-core/pkg/gdbutil"
	"lina-core/pkg/logger"
)

// Status represents user account status.
type Status int

const (
	// StatusNormal represents a normal user status.
	StatusNormal Status = 1

	// StatusDisabled represents a disabled user status.
	StatusDisabled Status = 0

	// DefaultAdminUsername is the username of the default admin user.
	DefaultAdminUsername = "admin"
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
	// Export generates an Excel file with user data based on IDs.
	Export(ctx context.Context, in ExportInput) (data []byte, err error)
	// Import reads an Excel file and creates users from it.
	Import(ctx context.Context, fileReader io.Reader) (result *ImportResult, err error)
	// GenerateImportTemplate creates an Excel template for user import.
	GenerateImportTemplate() (data []byte, err error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	authSvc   auth.Service
	bizCtxSvc bizctx.Service
	deptSvc   dept.Service
	roleSvc   role.Service // Role service
}

// New creates and returns a new Service instance.
func New() Service {
	return &serviceImpl{
		authSvc:   auth.New(),
		bizCtxSvc: bizctx.New(),
		deptSvc:   dept.New(),
		roleSvc:   role.New(),
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
	BeginTime      string // Creation time start
	EndTime        string // Creation time end
	OrderBy        string // Sort field
	OrderDirection string // Sort direction: asc/desc
}

// ListOutputItem defines a single item in list output with dept info.
type ListOutputItem struct {
	SysUser   *entity.SysUser // User entity
	DeptId    int             // Department ID
	DeptName  string          // Department name
	RoleIds   []int           // Role ID list
	RoleNames []string        // Role name list
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*ListOutputItem // User list
	Total int               // Total count
}

// List queries user list with pagination and filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysUser.Columns()
		m    = dao.SysUser.Ctx(ctx)
	)

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
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.CreatedAt, in.BeginTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(cols.CreatedAt, in.EndTime)
	}

	// Filter by dept via association table
	if in.DeptId != nil {
		if *in.DeptId == 0 {
			// Unassigned: users NOT in sys_user_dept
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

	// Batch query user-dept associations
	udCols := dao.SysUserDept.Columns()
	var userDepts []*entity.SysUserDept
	err = dao.SysUserDept.Ctx(ctx).
		WhereIn(udCols.UserId, userIds).
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}

	// Build userId -> deptId map
	userDeptMap := make(map[int]int)
	deptIds := make([]int, 0)
	for _, ud := range userDepts {
		userDeptMap[ud.UserId] = ud.DeptId
		deptIds = append(deptIds, ud.DeptId)
	}

	// Batch query dept info
	var depts []*entity.SysDept
	if len(deptIds) > 0 {
		err = dao.SysDept.Ctx(ctx).
			WhereIn(dao.SysDept.Columns().Id, deptIds).
			Scan(&depts)
		if err != nil {
			return nil, err
		}
	}

	// Build deptId -> deptName map
	deptNameMap := make(map[int]string)
	for _, d := range depts {
		deptNameMap[d.Id] = d.Name
	}

	// Build user-role associations
	urCols := dao.SysUserRole.Columns()
	var userRoles []*entity.SysUserRole
	err = dao.SysUserRole.Ctx(ctx).
		WhereIn(urCols.UserId, userIds).
		Scan(&userRoles)
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
		err = dao.SysRole.Ctx(ctx).
			WhereIn(roleCols.Id, allRoleIds).
			Scan(&roles)
		if err != nil {
			return nil, err
		}
	}

	// Build roleId -> roleName map
	roleNameMap := make(map[int]string)
	for _, r := range roles {
		roleNameMap[r.Id] = r.Name
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
		items = append(items, item)
	}

	return &ListOutput{
		List:  items,
		Total: total,
	}, nil
}

// GetUserIdsByDeptId returns user IDs associated with a dept and all its descendants.
func (s *serviceImpl) GetUserIdsByDeptId(ctx context.Context, deptId int) ([]int, error) {
	// Use shared method from dept service to get dept and descendant IDs
	deptIds, err := s.deptSvc.GetDeptAndDescendantIds(ctx, deptId)
	if err != nil {
		return nil, err
	}

	// Query users belonging to any of these depts
	var userDepts []*entity.SysUserDept
	err = dao.SysUserDept.Ctx(ctx).
		WhereIn(dao.SysUserDept.Columns().DeptId, deptIds).
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}
	// Deduplicate user IDs (a user could belong to multiple depts in the subtree)
	seen := make(map[int]struct{})
	ids := make([]int, 0, len(userDepts))
	for _, ud := range userDepts {
		if _, ok := seen[ud.UserId]; !ok {
			seen[ud.UserId] = struct{}{}
			ids = append(ids, ud.UserId)
		}
	}
	return ids, nil
}

// GetAllAssignedUserIds returns all user IDs that have a dept association.
func (s *serviceImpl) GetAllAssignedUserIds(ctx context.Context) ([]int, error) {
	var userDepts []*entity.SysUserDept
	err := dao.SysUserDept.Ctx(ctx).
		Fields(dao.SysUserDept.Columns().UserId).
		Distinct().
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(userDepts))
	for _, ud := range userDepts {
		ids = append(ids, ud.UserId)
	}
	return ids, nil
}

// GetUserDeptInfo returns the dept ID and name for a user.
func (s *serviceImpl) GetUserDeptInfo(ctx context.Context, userId int) (int, string, error) {
	var userDept *entity.SysUserDept
	err := dao.SysUserDept.Ctx(ctx).
		Where(dao.SysUserDept.Columns().UserId, userId).
		Scan(&userDept)
	if err != nil || userDept == nil {
		return 0, "", err
	}
	var dept *entity.SysDept
	err = dao.SysDept.Ctx(ctx).
		Where(dao.SysDept.Columns().Id, userDept.DeptId).
		Scan(&dept)
	if err != nil || dept == nil {
		return 0, "", err
	}
	return dept.Id, dept.Name, nil
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Username string // Username
	Password string // Password
	Nickname string // Nickname
	Email    string // Email
	Phone    string // Phone number
	Sex      int    // Gender: 0=Unknown 1=Male 2=Female
	Status   Status // Status: StatusNormal=Normal StatusDisabled=Disabled
	Remark   string // Remark
	DeptId   *int   // Department ID
	PostIds  []int  // Post ID list
	RoleIds  []int  // Role ID list
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
		return 0, gerror.New("用户名已存在")
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
		}).InsertAndGetId()
		if err != nil {
			return err
		}

		userId = int(id)

		// Save dept association
		if in.DeptId != nil && *in.DeptId > 0 {
			_, err = dao.SysUserDept.Ctx(ctx).Data(do.SysUserDept{
				UserId: userId,
				DeptId: *in.DeptId,
			}).Insert()
			if err != nil {
				return err
			}
		}

		// Save post associations
		for _, postId := range in.PostIds {
			_, err = dao.SysUserPost.Ctx(ctx).Data(do.SysUserPost{
				UserId: userId,
				PostId: postId,
			}).Insert()
			if err != nil {
				return err
			}
		}

		// Save role associations
		for _, roleId := range in.RoleIds {
			_, err = dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
				UserId: userId,
				RoleId: roleId,
			}).Insert()
			if err != nil {
				return err
			}
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
		return nil, gerror.New("用户不存在")
	}
	return user, nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id       int     // User ID
	Username *string // Username
	Password *string // Password
	Nickname *string // Nickname
	Email    *string // Email
	Phone    *string // Phone number
	Sex      *int    // Gender: 0=Unknown 1=Male 2=Female
	Status   *int    // Status: 1=Normal 0=Disabled
	Remark   *string // Remark
	DeptId   *int    // Department ID
	PostIds  []int   // Post ID list
	RoleIds  []int   // Role ID list
}

// Update updates user information with transaction support.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	// Cannot edit self via admin panel
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx != nil && bizCtx.UserId == in.Id {
		return gerror.New("不能编辑当前登录用户")
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

	// Use transaction to ensure atomicity
	err := dao.SysUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Update user
		_, err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: in.Id}).Data(data).Update()
		if err != nil {
			return err
		}

		// Update dept association (delete and re-insert)
		if in.DeptId != nil {
			_, err = dao.SysUserDept.Ctx(ctx).Where(dao.SysUserDept.Columns().UserId, in.Id).Delete()
			if err != nil {
				logger.Warningf(ctx, "failed to delete user dept association: %v", err)
			}
			if *in.DeptId > 0 {
				_, err = dao.SysUserDept.Ctx(ctx).Data(do.SysUserDept{
					UserId: in.Id,
					DeptId: *in.DeptId,
				}).Insert()
				if err != nil {
					return err
				}
			}
		}

		// Update post associations (delete and re-insert)
		if in.PostIds != nil {
			_, err = dao.SysUserPost.Ctx(ctx).Where(dao.SysUserPost.Columns().UserId, in.Id).Delete()
			if err != nil {
				logger.Warningf(ctx, "failed to delete user post association: %v", err)
			}
			for _, postId := range in.PostIds {
				_, err = dao.SysUserPost.Ctx(ctx).Data(do.SysUserPost{
					UserId: in.Id,
					PostId: postId,
				}).Insert()
				if err != nil {
					return err
				}
			}
		}

		// Update role associations (delete and re-insert)
		if in.RoleIds != nil {
			_, err = dao.SysUserRole.Ctx(ctx).Where(dao.SysUserRole.Columns().UserId, in.Id).Delete()
			if err != nil {
				logger.Warningf(ctx, "failed to delete user role association: %v", err)
			}
			for _, roleId := range in.RoleIds {
				_, err = dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
					UserId: in.Id,
					RoleId: roleId,
				}).Insert()
				if err != nil {
					return err
				}
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
	// Cannot delete default admin
	user, err := s.GetById(ctx, id)
	if err != nil {
		return err
	}
	if user.Username == DefaultAdminUsername {
		return gerror.New("不能删除默认管理员")
	}

	// Cannot delete self
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx != nil && bizCtx.UserId == id {
		return gerror.New("不能删除当前登录用户")
	}

	// Soft delete using GoFrame's auto soft-delete feature
	_, err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: id}).
		Delete()
	if err != nil {
		return err
	}

	// Clean up dept, post and role associations (log errors but don't fail)
	if _, err := dao.SysUserDept.Ctx(ctx).Where(dao.SysUserDept.Columns().UserId, id).Delete(); err != nil {
		logger.Warningf(ctx, "failed to delete user dept association for user %d: %v", id, err)
	}
	if _, err := dao.SysUserPost.Ctx(ctx).Where(dao.SysUserPost.Columns().UserId, id).Delete(); err != nil {
		logger.Warningf(ctx, "failed to delete user post association for user %d: %v", id, err)
	}
	if _, err := dao.SysUserRole.Ctx(ctx).Where(dao.SysUserRole.Columns().UserId, id).Delete(); err != nil {
		logger.Warningf(ctx, "failed to delete user role association for user %d: %v", id, err)
	}

	s.roleSvc.NotifyAccessTopologyChanged(ctx)
	return nil
}

// UpdateStatus updates user status.
func (s *serviceImpl) UpdateStatus(ctx context.Context, id int, status Status) error {
	// Cannot disable self
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx != nil && bizCtx.UserId == id && status == StatusDisabled {
		return gerror.New("不能停用当前登录用户")
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
		return nil, gerror.New("未登录")
	}
	return s.GetById(ctx, bizCtx.UserId)
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
		return gerror.New("未登录")
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
		return gerror.New("未登录")
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
	var userPosts []*entity.SysUserPost
	err := dao.SysUserPost.Ctx(ctx).
		Where(dao.SysUserPost.Columns().UserId, userId).
		Scan(&userPosts)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(userPosts))
	for _, up := range userPosts {
		ids = append(ids, up.PostId)
	}
	return ids, nil
}

// GetUserRoleIds returns the role IDs associated with a user.
func (s *serviceImpl) GetUserRoleIds(ctx context.Context, userId int) ([]int, error) {
	var userRoles []*entity.SysUserRole
	err := dao.SysUserRole.Ctx(ctx).
		Where(dao.SysUserRole.Columns().UserId, userId).
		Scan(&userRoles)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(userRoles))
	for _, ur := range userRoles {
		ids = append(ids, ur.RoleId)
	}
	return ids, nil
}
