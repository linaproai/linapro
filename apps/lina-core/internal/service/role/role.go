// Package role implements role management, permission lookup, and shared access
// context caching for the Lina core host service.
package role

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/config"
	"lina-core/internal/service/kvcache"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/logger"
)

// Service defines the role service contract.
type Service interface {
	// List queries role list with pagination.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById retrieves role by ID.
	GetById(ctx context.Context, id int) (*entity.SysRole, error)
	// GetDetail retrieves role detail with menu IDs.
	GetDetail(ctx context.Context, id int) (*GetDetailOutput, error)
	// Create creates a new role.
	Create(ctx context.Context, in CreateInput) (int, error)
	// Update updates role information.
	Update(ctx context.Context, in UpdateInput) error
	// Delete deletes a role.
	Delete(ctx context.Context, id int) error
	// UpdateStatus updates role status.
	UpdateStatus(ctx context.Context, id int, status int) error
	// GetOptions returns role options for dropdown.
	GetOptions(ctx context.Context) ([]*OptionItem, error)
	// GetUsers queries users assigned to a role.
	GetUsers(ctx context.Context, in GetUsersInput) (*GetUsersOutput, error)
	// AssignUsers assigns users to a role.
	AssignUsers(ctx context.Context, roleId int, userIds []int) error
	// UnassignUser removes user from a role.
	UnassignUser(ctx context.Context, roleId int, userId int) error
	// UnassignUsers removes multiple users from a role.
	UnassignUsers(ctx context.Context, roleId int, userIds []int) error
	// GetUserRoleIds returns role IDs for a user.
	GetUserRoleIds(ctx context.Context, userId int) ([]int, error)
	// GetUserRoles returns role entities for a user.
	GetUserRoles(ctx context.Context, userId int) ([]*entity.SysRole, error)
	// GetUserRoleNames returns role names for a user.
	GetUserRoleNames(ctx context.Context, userId int) ([]string, error)
	// GetUserMenuIds returns menu IDs accessible by a user through their roles.
	GetUserMenuIds(ctx context.Context, userId int) ([]int, error)
	// GetUserPermissions returns effective menu and button permission strings for a user.
	GetUserPermissions(ctx context.Context, userId int) ([]string, error)
	// IsSuperAdmin checks if user is a super admin (has admin role).
	IsSuperAdmin(ctx context.Context, userId int) bool
	// PrimeTokenAccessContext preloads the access context cache for one freshly issued login token.
	PrimeTokenAccessContext(
		ctx context.Context,
		tokenID string,
		userID int,
	) (*UserAccessContext, error)
	// InvalidateTokenAccessContext removes the cached access context bound to one token.
	InvalidateTokenAccessContext(ctx context.Context, tokenID string)
	// InvalidateUserAccessContexts removes all cached access contexts bound to one user.
	InvalidateUserAccessContexts(ctx context.Context, userID int)
	// MarkAccessTopologyChanged bumps the shared permission topology revision and clears local token caches.
	MarkAccessTopologyChanged(ctx context.Context) error
	// NotifyAccessTopologyChanged best-effort refreshes the shared permission topology revision.
	NotifyAccessTopologyChanged(ctx context.Context)
	// GetUserAccessContext loads the user's roles, menus, and permissions with token-aware caching when available.
	GetUserAccessContext(ctx context.Context, userId int) (*UserAccessContext, error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc  bizctx.Service
	configSvc  config.Service
	kvCacheSvc kvcache.Service
	pluginSvc  pluginsvc.Service
}

func New() Service {
	return &serviceImpl{
		bizCtxSvc:  bizctx.New(),
		configSvc:  config.New(),
		kvCacheSvc: kvcache.New(),
		pluginSvc:  pluginsvc.New(),
	}
}

type ListInput struct {
	Name   string
	Key    string
	Status *int
	Page   int
	Size   int
}

type ListOutput struct {
	List  []*RoleItem // Role list
	Total int         // Total count
}

// RoleItem represents a role in the list response.
type RoleItem struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	Sort      int    `json:"sort"`
	DataScope int    `json:"dataScope"`
	Status    int    `json:"status"`
	Remark    string `json:"remark"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// List queries role list with pagination.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysRole.Columns()
		m    = dao.SysRole.Ctx(ctx)
	)

	// Apply filters
	if in.Name != "" {
		m = m.WhereLike(cols.Name, "%"+in.Name+"%")
	}
	if in.Key != "" {
		m = m.WhereLike(cols.Key, "%"+in.Key+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Apply pagination
	offset := (in.Page - 1) * in.Size
	var roles []*entity.SysRole
	err = m.OrderAsc(cols.Sort).
		Limit(offset, in.Size).
		Scan(&roles)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	list := make([]*RoleItem, 0, len(roles))
	for _, r := range roles {
		createdAt := ""
		if r.CreatedAt != nil {
			createdAt = r.CreatedAt.String()
		}
		updatedAt := ""
		if r.UpdatedAt != nil {
			updatedAt = r.UpdatedAt.String()
		}
		list = append(list, &RoleItem{
			Id:        r.Id,
			Name:      r.Name,
			Key:       r.Key,
			Sort:      r.Sort,
			DataScope: r.DataScope,
			Status:    r.Status,
			Remark:    r.Remark,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	return &ListOutput{
		List:  list,
		Total: total,
	}, nil
}

// GetById retrieves role by ID.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*entity.SysRole, error) {
	var role *entity.SysRole
	err := dao.SysRole.Ctx(ctx).
		Where(do.SysRole{Id: id}).
		Scan(&role)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, gerror.New("角色不存在")
	}
	return role, nil
}

// GetDetailOutput defines output for GetDetail function.
type GetDetailOutput struct {
	Role    *entity.SysRole
	MenuIds []int
}

// GetDetail retrieves role detail with menu IDs.
func (s *serviceImpl) GetDetail(ctx context.Context, id int) (*GetDetailOutput, error) {
	// Get role
	role, err := s.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get associated menu IDs
	rmCols := dao.SysRoleMenu.Columns()
	var roleMenus []*entity.SysRoleMenu
	err = dao.SysRoleMenu.Ctx(ctx).
		Where(rmCols.RoleId, id).
		Scan(&roleMenus)
	if err != nil {
		return nil, err
	}

	menuIds := make([]int, 0, len(roleMenus))
	for _, rm := range roleMenus {
		menuIds = append(menuIds, rm.MenuId)
	}

	return &GetDetailOutput{
		Role:    role,
		MenuIds: menuIds,
	}, nil
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Name      string
	Key       string
	Sort      int
	DataScope int
	Status    int
	Remark    string
	MenuIds   []int
}

// Create creates a new role.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int, error) {
	// Check name uniqueness
	if err := s.checkNameUnique(ctx, in.Name, 0); err != nil {
		return 0, err
	}

	// Check key uniqueness
	if err := s.checkKeyUnique(ctx, in.Key, 0); err != nil {
		return 0, err
	}

	// Use transaction
	var roleId int64
	err := dao.SysRole.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Insert role (GoFrame auto-fills created_at and updated_at)
		id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
			Name:      in.Name,
			Key:       in.Key,
			Sort:      in.Sort,
			DataScope: in.DataScope,
			Status:    in.Status,
			Remark:    in.Remark,
		}).InsertAndGetId()
		if err != nil {
			return err
		}
		roleId = id

		// Insert role-menu associations
		if len(in.MenuIds) > 0 {
			for _, menuId := range in.MenuIds {
				_, err = dao.SysRoleMenu.Ctx(ctx).Data(do.SysRoleMenu{
					RoleId: int(roleId),
					MenuId: menuId,
				}).Insert()
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}
	s.NotifyAccessTopologyChanged(ctx)

	return int(roleId), nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id        int
	Name      string
	Key       string
	Sort      *int
	DataScope *int
	Status    *int
	Remark    *string
	MenuIds   []int
}

// Update updates role information.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	// Check role exists
	_, err := s.GetById(ctx, in.Id)
	if err != nil {
		return err
	}

	// Check name uniqueness (excluding self)
	if err := s.checkNameUnique(ctx, in.Name, in.Id); err != nil {
		return err
	}

	// Check key uniqueness (excluding self)
	if err := s.checkKeyUnique(ctx, in.Key, in.Id); err != nil {
		return err
	}

	// Use transaction
	err = dao.SysRole.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Update role
		data := do.SysRole{
			Name: in.Name,
			Key:  in.Key,
		}
		if in.Sort != nil {
			data.Sort = *in.Sort
		}
		if in.DataScope != nil {
			data.DataScope = *in.DataScope
		}
		if in.Status != nil {
			data.Status = *in.Status
		}
		if in.Remark != nil {
			data.Remark = *in.Remark
		}

		_, err = dao.SysRole.Ctx(ctx).Where(do.SysRole{Id: in.Id}).Data(data).Update()
		if err != nil {
			return err
		}

		// Delete old role-menu associations
		rmCols := dao.SysRoleMenu.Columns()
		_, err = dao.SysRoleMenu.Ctx(ctx).
			Where(rmCols.RoleId, in.Id).
			Delete()
		if err != nil {
			return err
		}

		// Insert new role-menu associations
		if len(in.MenuIds) > 0 {
			for _, menuId := range in.MenuIds {
				_, err = dao.SysRoleMenu.Ctx(ctx).Data(do.SysRoleMenu{
					RoleId: in.Id,
					MenuId: menuId,
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
	s.NotifyAccessTopologyChanged(ctx)
	return nil
}

// Delete deletes a role.
func (s *serviceImpl) Delete(ctx context.Context, id int) error {
	// Check role exists
	_, err := s.GetById(ctx, id)
	if err != nil {
		return err
	}

	// Use transaction
	err = dao.SysRole.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Delete role-menu associations
		rmCols := dao.SysRoleMenu.Columns()
		_, err = dao.SysRoleMenu.Ctx(ctx).
			Where(rmCols.RoleId, id).
			Delete()
		if err != nil {
			logger.Warningf(ctx, "failed to delete role-menu associations: %v", err)
		}

		// Delete user-role associations
		urCols := dao.SysUserRole.Columns()
		_, err = dao.SysUserRole.Ctx(ctx).
			Where(urCols.RoleId, id).
			Delete()
		if err != nil {
			logger.Warningf(ctx, "failed to delete user-role associations: %v", err)
		}

		// Delete role
		_, err = dao.SysRole.Ctx(ctx).
			Where(do.SysRole{Id: id}).
			Delete()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	s.NotifyAccessTopologyChanged(ctx)
	return nil
}

// UpdateStatus updates role status.
func (s *serviceImpl) UpdateStatus(ctx context.Context, id int, status int) error {
	// Check role exists
	_, err := s.GetById(ctx, id)
	if err != nil {
		return err
	}

	_, err = dao.SysRole.Ctx(ctx).
		Where(do.SysRole{Id: id}).
		Data(do.SysRole{Status: status}).
		Update()
	if err != nil {
		return err
	}
	s.NotifyAccessTopologyChanged(ctx)
	return nil
}

// OptionItem represents a role option.
type OptionItem struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

// GetOptions returns role options for dropdown.
func (s *serviceImpl) GetOptions(ctx context.Context) ([]*OptionItem, error) {
	var roles []*entity.SysRole
	cols := dao.SysRole.Columns()
	err := dao.SysRole.Ctx(ctx).
		Where(cols.Status, 1).
		OrderAsc(cols.Sort).
		Scan(&roles)
	if err != nil {
		return nil, err
	}

	list := make([]*OptionItem, 0, len(roles))
	for _, r := range roles {
		list = append(list, &OptionItem{
			Id:   r.Id,
			Name: r.Name,
			Key:  r.Key,
		})
	}

	return list, nil
}

// RoleUserItem represents a user assigned to a role.
type RoleUserItem struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// GetUsersInput defines input for GetUsers function.
type GetUsersInput struct {
	RoleId   int
	Username string
	Phone    string
	Status   *int
	Page     int
	Size     int
}

// GetUsersOutput defines output for GetUsers function.
type GetUsersOutput struct {
	List  []*RoleUserItem
	Total int
}

// GetUsers queries users assigned to a role.
func (s *serviceImpl) GetUsers(ctx context.Context, in GetUsersInput) (*GetUsersOutput, error) {
	// Check role exists
	_, err := s.GetById(ctx, in.RoleId)
	if err != nil {
		return nil, err
	}

	// Get user IDs for this role
	urCols := dao.SysUserRole.Columns()
	var userRoles []*entity.SysUserRole
	err = dao.SysUserRole.Ctx(ctx).
		Where(urCols.RoleId, in.RoleId).
		Scan(&userRoles)
	if err != nil {
		return nil, err
	}

	if len(userRoles) == 0 {
		return &GetUsersOutput{
			List:  []*RoleUserItem{},
			Total: 0,
		}, nil
	}

	userIds := make([]int, 0, len(userRoles))
	for _, ur := range userRoles {
		userIds = append(userIds, ur.UserId)
	}

	// Query users with filters
	userCols := dao.SysUser.Columns()
	m := dao.SysUser.Ctx(ctx).WhereIn(userCols.Id, userIds)

	if in.Username != "" {
		m = m.WhereLike(userCols.Username, "%"+in.Username+"%")
	}
	if in.Phone != "" {
		m = m.WhereLike(userCols.Phone, "%"+in.Phone+"%")
	}
	if in.Status != nil {
		m = m.Where(userCols.Status, *in.Status)
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Apply pagination
	offset := (in.Page - 1) * in.Size
	var users []*entity.SysUser
	err = m.OrderDesc(userCols.Id).
		Limit(offset, in.Size).
		Scan(&users)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	list := make([]*RoleUserItem, 0, len(users))
	for _, u := range users {
		createdAt := ""
		if u.CreatedAt != nil {
			createdAt = u.CreatedAt.String()
		}
		list = append(list, &RoleUserItem{
			Id:        u.Id,
			Username:  u.Username,
			Nickname:  u.Nickname,
			Email:     u.Email,
			Phone:     u.Phone,
			Status:    u.Status,
			CreatedAt: createdAt,
		})
	}

	return &GetUsersOutput{
		List:  list,
		Total: total,
	}, nil
}

// AssignUsers assigns users to a role.
func (s *serviceImpl) AssignUsers(ctx context.Context, roleId int, userIds []int) error {
	// Check role exists
	_, err := s.GetById(ctx, roleId)
	if err != nil {
		return err
	}

	// Get existing user-role associations
	urCols := dao.SysUserRole.Columns()
	var existingRoles []*entity.SysUserRole
	err = dao.SysUserRole.Ctx(ctx).
		Where(urCols.RoleId, roleId).
		Scan(&existingRoles)
	if err != nil {
		return err
	}

	existingUserIds := make(map[int]bool)
	for _, ur := range existingRoles {
		existingUserIds[ur.UserId] = true
	}

	// Insert new associations (skip existing)
	for _, userId := range userIds {
		if existingUserIds[userId] {
			continue
		}
		_, err = dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
			UserId: userId,
			RoleId: roleId,
		}).Insert()
		if err != nil {
			logger.Warningf(ctx, "failed to assign user %d to role %d: %v", userId, roleId, err)
		}
	}

	s.NotifyAccessTopologyChanged(ctx)
	return nil
}

// UnassignUser removes user from a role.
func (s *serviceImpl) UnassignUser(ctx context.Context, roleId int, userId int) error {
	// Check role exists
	_, err := s.GetById(ctx, roleId)
	if err != nil {
		return err
	}

	urCols := dao.SysUserRole.Columns()
	_, err = dao.SysUserRole.Ctx(ctx).
		Where(urCols.RoleId, roleId).
		Where(urCols.UserId, userId).
		Delete()
	if err != nil {
		return err
	}
	s.NotifyAccessTopologyChanged(ctx)
	return nil
}

// UnassignUsers removes multiple users from a role.
func (s *serviceImpl) UnassignUsers(ctx context.Context, roleId int, userIds []int) error {
	// Check role exists
	_, err := s.GetById(ctx, roleId)
	if err != nil {
		return err
	}

	urCols := dao.SysUserRole.Columns()
	_, err = dao.SysUserRole.Ctx(ctx).
		Where(urCols.RoleId, roleId).
		WhereIn(urCols.UserId, userIds).
		Delete()
	if err != nil {
		return err
	}
	s.NotifyAccessTopologyChanged(ctx)
	return nil
}

// checkNameUnique checks if the role name is unique.
func (s *serviceImpl) checkNameUnique(ctx context.Context, name string, excludeId int) error {
	cols := dao.SysRole.Columns()
	m := dao.SysRole.Ctx(ctx).Where(cols.Name, name)
	if excludeId > 0 {
		m = m.WhereNot(cols.Id, excludeId)
	}
	count, err := m.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("角色名称已存在")
	}
	return nil
}

// checkKeyUnique checks if the role key is unique.
func (s *serviceImpl) checkKeyUnique(ctx context.Context, key string, excludeId int) error {
	cols := dao.SysRole.Columns()
	m := dao.SysRole.Ctx(ctx).Where(cols.Key, key)
	if excludeId > 0 {
		m = m.WhereNot(cols.Id, excludeId)
	}
	count, err := m.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("权限字符已存在")
	}
	return nil
}

// GetUserRoleIds returns role IDs for a user.
func (s *serviceImpl) GetUserRoleIds(ctx context.Context, userId int) ([]int, error) {
	urCols := dao.SysUserRole.Columns()
	var userRoles []*entity.SysUserRole
	err := dao.SysUserRole.Ctx(ctx).
		Where(urCols.UserId, userId).
		Scan(&userRoles)
	if err != nil {
		return nil, err
	}

	roleIds := make([]int, 0, len(userRoles))
	for _, ur := range userRoles {
		roleIds = append(roleIds, ur.RoleId)
	}

	return roleIds, nil
}

// GetUserRoles returns role entities for a user.
func (s *serviceImpl) GetUserRoles(ctx context.Context, userId int) ([]*entity.SysRole, error) {
	roleIds, err := s.GetUserRoleIds(ctx, userId)
	if err != nil {
		return nil, err
	}
	return s.getUserRolesByRoleIds(ctx, roleIds)
}

// GetUserRoleNames returns role names for a user.
func (s *serviceImpl) GetUserRoleNames(ctx context.Context, userId int) ([]string, error) {
	roles, err := s.GetUserRoles(ctx, userId)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(roles))
	for _, r := range roles {
		names = append(names, r.Name)
	}

	return names, nil
}

// GetUserMenuIds returns menu IDs accessible by a user through their roles.
func (s *serviceImpl) GetUserMenuIds(ctx context.Context, userId int) ([]int, error) {
	roleIds, err := s.GetUserRoleIds(ctx, userId)
	if err != nil {
		return nil, err
	}
	return s.getUserMenuIdsByRoleIds(ctx, roleIds)
}

// GetUserPermissions returns effective menu and button permission strings for a user.
func (s *serviceImpl) GetUserPermissions(ctx context.Context, userId int) ([]string, error) {
	menuIds, err := s.GetUserMenuIds(ctx, userId)
	if err != nil {
		return nil, err
	}
	return s.getUserPermissionsByMenuIds(ctx, menuIds)
}

// IsSuperAdmin checks if user is a super admin (has admin role).
func (s *serviceImpl) IsSuperAdmin(ctx context.Context, userId int) bool {
	roleIds, err := s.GetUserRoleIds(ctx, userId)
	if err != nil {
		return false
	}

	// Check if user has admin role (roleId = 1)
	for _, roleId := range roleIds {
		if roleId == 1 {
			return true
		}
	}

	return false
}
