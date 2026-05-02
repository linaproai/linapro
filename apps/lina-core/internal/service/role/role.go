// Package role implements role management, permission lookup, and shared access
// context caching for the Lina core host service.
package role

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/pkg/bizerr"
)

const (
	// builtinAdminRoleKey identifies the protected built-in administrator role.
	builtinAdminRoleKey = "admin"
	// builtinAdminRoleNameI18n is the runtime i18n key for the built-in administrator role.
	builtinAdminRoleNameI18n = "role.builtin.admin.name"
	// builtinUserRoleKey identifies the protected built-in standard user role.
	builtinUserRoleKey = "user"
	// builtinUserRoleNameI18n is the runtime i18n key for the built-in standard user role.
	builtinUserRoleNameI18n = "role.builtin.user.name"
)

// PermissionMenuFilter defines the narrow dependency required by the role
// service to hide plugin-owned permission menus that are not currently active.
type PermissionMenuFilter interface {
	// FilterPermissionMenus returns only the permission menus that should remain
	// effective for the current host and plugin state.
	FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu
}

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
	// BatchDelete deletes multiple roles atomically.
	BatchDelete(ctx context.Context, ids []int) error
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
	// DisplayName returns the read-only display name for one role, localizing
	// protected built-in roles while preserving custom role names.
	DisplayName(ctx context.Context, role *entity.SysRole) string
	// GetUserMenuIds returns menu IDs accessible by a user through their roles.
	GetUserMenuIds(ctx context.Context, userId int) ([]int, error)
	// GetUserPermissions returns effective menu and button permission strings for a user.
	GetUserPermissions(ctx context.Context, userId int) ([]string, error)
	// IsSuperAdmin checks whether the user is the built-in admin account.
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
	// SyncAccessTopologyRevision synchronizes the process-local permission
	// topology revision and evicts stale token snapshots after cross-node changes.
	SyncAccessTopologyRevision(ctx context.Context) error
	// GetUserAccessContext loads the user's roles, menus, and permissions with token-aware caching when available.
	GetUserAccessContext(ctx context.Context, userId int) (*UserAccessContext, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc          bizctx.Service
	configSvc          config.Service
	i18nSvc            roleI18nTranslator
	kvCacheSvc         kvcache.Service
	permissionFilter   PermissionMenuFilter
	accessRevisionCtrl accessRevisionController
}

// New creates and returns a new role Service.
// Pass a non-nil permissionFilter when role permission calculation must respect
// plugin-owned permission menu visibility; pass nil to use the default no-op filter.
func New(permissionFilter PermissionMenuFilter) Service {
	var (
		bizCtxSvc  = bizctx.New()
		configSvc  = config.New()
		i18nSvc    = i18nsvc.New()
		kvCacheSvc = kvcache.New()
	)
	if permissionFilter == nil {
		permissionFilter = noopPermissionMenuFilter{}
	}

	return &serviceImpl{
		bizCtxSvc:        bizCtxSvc,
		configSvc:        configSvc,
		i18nSvc:          i18nSvc,
		kvCacheSvc:       kvCacheSvc,
		permissionFilter: permissionFilter,
		accessRevisionCtrl: newAccessRevisionController(
			configSvc.IsClusterEnabled(context.Background()),
			kvCacheSvc,
		),
	}
}

// roleI18nTranslator defines the narrow translation capability role needs.
type roleI18nTranslator interface {
	// Translate returns one runtime translation key with caller-provided fallback text.
	Translate(ctx context.Context, key string, fallback string) string
}

// noopPermissionMenuFilter keeps permission menus unchanged when no external
// plugin-aware filter is injected into the role service.
type noopPermissionMenuFilter struct{}

// FilterPermissionMenus returns the original menu slice unchanged.
func (noopPermissionMenuFilter) FilterPermissionMenus(_ context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return menus
}

// ListInput defines filters and pagination for role list queries.
type ListInput struct {
	Name   string
	Key    string
	Status *int
	Page   int
	Size   int
}

// ListOutput defines the paged role list result.
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
			Name:      s.DisplayName(ctx, r),
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

// DisplayName translates protected built-in role names in read-only display
// rows while keeping editable role records and custom roles unchanged.
func (s *serviceImpl) DisplayName(ctx context.Context, role *entity.SysRole) string {
	if role == nil {
		return ""
	}
	if s == nil || s.i18nSvc == nil {
		return role.Name
	}
	switch role.Key {
	case builtinAdminRoleKey:
		return s.i18nSvc.Translate(ctx, builtinAdminRoleNameI18n, role.Name)
	case builtinUserRoleKey:
		return s.i18nSvc.Translate(ctx, builtinUserRoleNameI18n, role.Name)
	default:
		return role.Name
	}
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
		return nil, bizerr.NewCode(CodeRoleNotFound)
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
		if err = insertRoleMenus(ctx, int(roleId), in.MenuIds); err != nil {
			return err
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
		if err = insertRoleMenus(ctx, in.Id, in.MenuIds); err != nil {
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

// insertRoleMenus inserts all role-menu associations for one role in a single batch.
func insertRoleMenus(ctx context.Context, roleID int, menuIDs []int) error {
	relations := buildRoleMenuRelations(roleID, menuIDs)
	if len(relations) == 0 {
		return nil
	}
	_, err := dao.SysRoleMenu.Ctx(ctx).Data(relations).Insert()
	return err
}

// buildRoleMenuRelations normalizes menu IDs into distinct role-menu rows.
func buildRoleMenuRelations(roleID int, menuIDs []int) []do.SysRoleMenu {
	if roleID <= 0 || len(menuIDs) == 0 {
		return []do.SysRoleMenu{}
	}
	seen := make(map[int]struct{}, len(menuIDs))
	relations := make([]do.SysRoleMenu, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		if menuID <= 0 {
			continue
		}
		if _, ok := seen[menuID]; ok {
			continue
		}
		seen[menuID] = struct{}{}
		relations = append(relations, do.SysRoleMenu{
			RoleId: roleID,
			MenuId: menuID,
		})
	}
	return relations
}

// Delete deletes a role.
func (s *serviceImpl) Delete(ctx context.Context, id int) error {
	if err := s.runRoleDeletionTransaction(ctx, []int{id}); err != nil {
		return err
	}
	s.NotifyAccessTopologyChanged(ctx)
	return nil
}

// BatchDelete deletes multiple roles atomically.
func (s *serviceImpl) BatchDelete(ctx context.Context, ids []int) error {
	normalizedIds := normalizeRoleDeleteIDs(ids)
	if len(normalizedIds) == 0 {
		return bizerr.NewCode(CodeRoleDeleteIdsRequired)
	}
	if err := s.runRoleDeletionTransaction(ctx, normalizedIds); err != nil {
		return err
	}
	s.NotifyAccessTopologyChanged(ctx)
	return nil
}

// runRoleDeletionTransaction validates and deletes roles with associations in one transaction.
func (s *serviceImpl) runRoleDeletionTransaction(ctx context.Context, ids []int) error {
	return dao.SysRole.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		for _, id := range ids {
			if err := s.ensureRoleDeleteAllowed(ctx, id); err != nil {
				return err
			}
		}
		for _, id := range ids {
			if err := s.deleteRoleRecordAndAssociations(ctx, id); err != nil {
				return err
			}
		}
		return nil
	})
}

// ensureRoleDeleteAllowed enforces built-in role deletion protection.
func (s *serviceImpl) ensureRoleDeleteAllowed(ctx context.Context, id int) error {
	role, err := s.GetById(ctx, id)
	if err != nil {
		return err
	}
	if role.Key == builtinAdminRoleKey {
		return bizerr.NewCode(CodeRoleBuiltinDeleteDenied)
	}
	return nil
}

// deleteRoleRecordAndAssociations soft-deletes one role and clears its associations.
func (s *serviceImpl) deleteRoleRecordAndAssociations(ctx context.Context, id int) error {
	rmCols := dao.SysRoleMenu.Columns()
	if _, err := dao.SysRoleMenu.Ctx(ctx).Where(rmCols.RoleId, id).Delete(); err != nil {
		return err
	}

	urCols := dao.SysUserRole.Columns()
	if _, err := dao.SysUserRole.Ctx(ctx).Where(urCols.RoleId, id).Delete(); err != nil {
		return err
	}

	_, err := dao.SysRole.Ctx(ctx).Where(do.SysRole{Id: id}).Delete()
	return err
}

// normalizeRoleDeleteIDs removes duplicate IDs while preserving request order.
func normalizeRoleDeleteIDs(ids []int) []int {
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

	err = dao.SysUserRole.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		// Get existing user-role associations.
		urCols := dao.SysUserRole.Columns()
		var existingRoles []*entity.SysUserRole
		if scanErr := dao.SysUserRole.Ctx(ctx).
			Where(urCols.RoleId, roleId).
			Scan(&existingRoles); scanErr != nil {
			return scanErr
		}

		existingUserIds := make(map[int]bool, len(existingRoles))
		for _, ur := range existingRoles {
			existingUserIds[ur.UserId] = true
		}

		newRelations := make([]do.SysUserRole, 0, len(userIds))
		for _, userId := range userIds {
			if existingUserIds[userId] {
				continue
			}
			existingUserIds[userId] = true
			newRelations = append(newRelations, do.SysUserRole{
				UserId: userId,
				RoleId: roleId,
			})
		}
		if len(newRelations) == 0 {
			return nil
		}
		_, insertErr := dao.SysUserRole.Ctx(ctx).Data(newRelations).Insert()
		return insertErr
	})
	if err != nil {
		return err
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
		return bizerr.NewCode(CodeRoleNameExists)
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
		return bizerr.NewCode(CodeRoleKeyExists)
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
		names = append(names, s.DisplayName(ctx, r))
	}

	return names, nil
}

// GetUserMenuIds returns menu IDs accessible by a user through their roles.
func (s *serviceImpl) GetUserMenuIds(ctx context.Context, userId int) ([]int, error) {
	accessContext, err := s.GetUserAccessContext(ctx, userId)
	if err != nil {
		return nil, err
	}
	if accessContext == nil {
		return []int{}, nil
	}
	return cloneSliceWithCopy(accessContext.MenuIds), nil
}

// GetUserPermissions returns effective menu and button permission strings for a user.
func (s *serviceImpl) GetUserPermissions(ctx context.Context, userId int) ([]string, error) {
	accessContext, err := s.GetUserAccessContext(ctx, userId)
	if err != nil {
		return nil, err
	}
	if accessContext == nil {
		return []string{}, nil
	}
	return cloneSliceWithCopy(accessContext.Permissions), nil
}

// IsSuperAdmin checks whether the user is the built-in admin account.
func (s *serviceImpl) IsSuperAdmin(ctx context.Context, userId int) bool {
	isSuperAdmin, err := s.isDefaultAdminUser(ctx, userId)
	if err != nil {
		return false
	}
	return isSuperAdmin
}
