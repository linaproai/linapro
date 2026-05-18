// role_impl.go implements role CRUD, authorization binding, menu access, and
// access-cache invalidation. It applies tenant and data-scope boundaries before
// reads or writes and keeps cache revision changes on the injected coordinator
// so role visibility stays consistent across runtime paths.

package role

import (
	"context"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/orgcap"
	pkgtenantcap "lina-core/pkg/tenantcap"

	"github.com/gogf/gf/v2/database/gdb"
)

// SetDataScopeService wires the shared data-scope service used by role user operations.
func (s *serviceImpl) SetDataScopeService(scopeSvc datascope.Service) {
	if s == nil {
		return
	}
	s.scopeSvc = scopeSvc
}

// FilterPermissionMenus returns the original menu slice unchanged.
func (noopPermissionMenuFilter) FilterPermissionMenus(_ context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return menus
}

// organizationCapabilityStateFromPermissionFilter reuses the plugin service
// dependency when it also exposes plugin enablement state.
func organizationCapabilityStateFromPermissionFilter(permissionFilter PermissionMenuFilter) OrganizationCapabilityState {
	if state, ok := permissionFilter.(OrganizationCapabilityState); ok {
		return state
	}
	if pluginState, ok := permissionFilter.(pluginEnablementState); ok {
		return pluginBackedOrganizationCapabilityState{pluginState: pluginState}
	}
	return nil
}

// Enabled reports whether org-center is enabled and the orgcap provider exists.
func (s pluginBackedOrganizationCapabilityState) Enabled(ctx context.Context) bool {
	return s.pluginState != nil &&
		s.pluginState.IsEnabled(ctx, orgcap.ProviderPluginID) &&
		orgcap.HasProvider()
}

// List queries role list with pagination.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysRole.Columns()
		m    = dao.SysRole.Ctx(ctx)
	)
	m = datascope.ApplyTenantScope(ctx, m, datascope.TenantColumn)

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
	model := dao.SysRole.Ctx(ctx).Where(do.SysRole{Id: id})
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	err := model.Scan(&role)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, bizerr.NewCode(CodeRoleNotFound)
	}
	return role, nil
}

// GetDetail retrieves role detail with menu IDs.
func (s *serviceImpl) GetDetail(ctx context.Context, id int) (*GetDetailOutput, error) {
	// Get role
	role, err := s.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get associated menu IDs
	var roleMenus []*entity.SysRoleMenu
	err = roleMenuModelForCurrentTenant(ctx, id).Scan(&roleMenus)
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

// Create creates a new role.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int, error) {
	if err := s.ensureRoleDataScopeAllowed(ctx, in.DataScope); err != nil {
		return 0, err
	}
	ownership := currentRoleOwnership(ctx)
	if err := ensureTenantRoleDataScopeBoundary(ownership.TenantID, in.DataScope); err != nil {
		return 0, err
	}

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
			TenantId:  ownership.TenantID,
		}).InsertAndGetId()
		if err != nil {
			return err
		}
		roleId = id

		// Insert role-menu associations
		if err = insertRoleMenus(ctx, int(roleId), in.MenuIds, ownership.TenantID); err != nil {
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

// Update updates role information.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	if in.DataScope != nil {
		if err := s.ensureRoleDataScopeAllowed(ctx, *in.DataScope); err != nil {
			return err
		}
	}
	// Check role exists
	role, err := s.GetById(ctx, in.Id)
	if err != nil {
		return err
	}
	ownership := roleOwnershipFromRole(role)
	if in.DataScope != nil {
		if err := ensureTenantRoleDataScopeBoundary(ownership.TenantID, *in.DataScope); err != nil {
			return err
		}
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
		_, err = roleMenuModelForCurrentTenant(ctx, in.Id).Delete()
		if err != nil {
			return err
		}

		// Insert new role-menu associations
		if err = insertRoleMenus(ctx, in.Id, in.MenuIds, ownership.TenantID); err != nil {
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

// currentRoleOwnership derives new-role ownership from the active tenant
// context. Platform context creates platform roles; tenant context creates
// tenant-local roles.
func currentRoleOwnership(ctx context.Context) roleOwnership {
	tenantID := datascope.CurrentTenantID(ctx)
	return roleOwnership{
		TenantID: tenantID,
	}
}

// roleOwnershipFromRole returns persisted role ownership metadata.
func roleOwnershipFromRole(role *entity.SysRole) roleOwnership {
	if role == nil {
		return currentRoleOwnership(context.Background())
	}
	return roleOwnership{
		TenantID: role.TenantId,
	}
}

// ensureTenantRoleDataScopeBoundary rejects global data scope on tenant-local roles.
func ensureTenantRoleDataScopeBoundary(tenantID int, dataScope int) error {
	if tenantID != datascope.PlatformTenantID && dataScope == roleDataScopeAll {
		return bizerr.NewCode(CodeTenantRoleAllDataScopeForbidden)
	}
	return nil
}

// ensureRoleDataScopeAllowed rejects organization-dependent role scopes when
// the organization management capability is not enabled.
func (s *serviceImpl) ensureRoleDataScopeAllowed(ctx context.Context, dataScope int) error {
	switch dataScope {
	case roleDataScopeAll, roleDataScopeTenant, roleDataScopeDept, roleDataScopeSelf:
	default:
		return bizerr.NewCode(CodeRoleDataScopeUnsupported, bizerr.P("scope", dataScope))
	}
	if dataScope != roleDataScopeDept {
		return nil
	}
	if s != nil && s.orgCapabilityState != nil && s.orgCapabilityState.Enabled(ctx) {
		return nil
	}
	return bizerr.NewCode(CodeRoleDataScopeDeptUnavailable)
}

// insertRoleMenus inserts all role-menu associations for one role in a single
// tenant boundary in a single batch.
func insertRoleMenus(ctx context.Context, roleID int, menuIDs []int, tenantID int) error {
	relations := buildRoleMenuRelations(roleID, menuIDs, tenantID)
	if len(relations) == 0 {
		return nil
	}
	_, err := dao.SysRoleMenu.Ctx(ctx).Data(relations).Insert()
	return err
}

// buildRoleMenuRelations normalizes menu IDs into distinct role-menu rows.
func buildRoleMenuRelations(roleID int, menuIDs []int, tenantID int) []do.SysRoleMenu {
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
			RoleId:   roleID,
			MenuId:   menuID,
			TenantId: tenantID,
		})
	}
	return relations
}

// roleMenuModelForCurrentTenant returns role-menu relation rows that belong to
// the active tenant. Platform context intentionally keeps the platform/global
// view.
func roleMenuModelForCurrentTenant(ctx context.Context, roleID int) *gdb.Model {
	rmCols := dao.SysRoleMenu.Columns()
	model := dao.SysRoleMenu.Ctx(ctx).Where(rmCols.RoleId, roleID)
	return datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
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
	if _, err := roleMenuModelForCurrentTenant(ctx, id).Delete(); err != nil {
		return err
	}

	userRoleModel := dao.SysUserRole.Ctx(ctx).Where(dao.SysUserRole.Columns().RoleId, id)
	userRoleModel = datascope.ApplyTenantScope(ctx, userRoleModel, datascope.TenantColumn)
	if _, err := userRoleModel.Delete(); err != nil {
		return err
	}

	roleModel := dao.SysRole.Ctx(ctx).Where(do.SysRole{Id: id})
	roleModel = datascope.ApplyTenantScope(ctx, roleModel, datascope.TenantColumn)
	_, err := roleModel.Delete()
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

	model := dao.SysRole.Ctx(ctx).Where(do.SysRole{Id: id})
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	_, err = model.Data(do.SysRole{Status: status}).Update()
	if err != nil {
		return err
	}
	s.NotifyAccessTopologyChanged(ctx)
	return nil
}

// GetOptions returns role options for dropdown.
func (s *serviceImpl) GetOptions(ctx context.Context) ([]*OptionItem, error) {
	var roles []*entity.SysRole
	cols := dao.SysRole.Columns()
	model := dao.SysRole.Ctx(ctx).
		Where(cols.Status, 1).
		OrderAsc(cols.Sort)
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	err := model.Scan(&roles)
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
	userRoleModel := dao.SysUserRole.Ctx(ctx).Where(urCols.RoleId, in.RoleId)
	userRoleModel = datascope.ApplyTenantScope(ctx, userRoleModel, datascope.TenantColumn)
	err = userRoleModel.Scan(&userRoles)
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
	m, scopeEmpty, err := s.applyRoleUserDataScope(ctx, m)
	if err != nil {
		return nil, err
	}
	if scopeEmpty {
		return &GetUsersOutput{List: []*RoleUserItem{}, Total: 0}, nil
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
	normalizedUserIDs := normalizeRoleAssignmentUserIDs(userIds)
	// Check role exists
	role, err := s.GetById(ctx, roleId)
	if err != nil {
		return err
	}
	if err = ensureRoleAssignmentBoundary(ctx, role); err != nil {
		return err
	}
	if err = s.ensureRoleUsersVisible(ctx, normalizedUserIDs); err != nil {
		return err
	}
	if err = s.ensureRoleAssignmentUsersMatchRoleBoundary(ctx, role, normalizedUserIDs); err != nil {
		return err
	}
	tenantID := role.TenantId

	err = dao.SysUserRole.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		// Get existing user-role associations.
		urCols := dao.SysUserRole.Columns()
		var existingRoles []*entity.SysUserRole
		existingModel := dao.SysUserRole.Ctx(ctx).Where(urCols.RoleId, roleId)
		existingModel = datascope.ApplyTenantScope(ctx, existingModel, datascope.TenantColumn)
		if scanErr := existingModel.Scan(&existingRoles); scanErr != nil {
			return scanErr
		}

		existingUserIds := make(map[int]bool, len(existingRoles))
		for _, ur := range existingRoles {
			existingUserIds[ur.UserId] = true
		}

		newRelations := make([]do.SysUserRole, 0, len(normalizedUserIDs))
		for _, userId := range normalizedUserIDs {
			if existingUserIds[userId] {
				continue
			}
			existingUserIds[userId] = true
			newRelations = append(newRelations, do.SysUserRole{
				UserId:   userId,
				RoleId:   roleId,
				TenantId: tenantID,
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

// normalizeRoleAssignmentUserIDs removes invalid and duplicate assignment
// targets while preserving request order.
func normalizeRoleAssignmentUserIDs(userIDs []int) []int {
	if len(userIDs) == 0 {
		return []int{}
	}
	normalizedIDs := make([]int, 0, len(userIDs))
	seen := make(map[int]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID <= 0 {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		normalizedIDs = append(normalizedIDs, userID)
	}
	return normalizedIDs
}

// ensureRoleAssignmentBoundary verifies role ownership matches the current
// tenant boundary before user-role rows are added.
func ensureRoleAssignmentBoundary(ctx context.Context, role *entity.SysRole) error {
	if role == nil {
		return bizerr.NewCode(CodeRoleNotFound)
	}
	tenantID := datascope.CurrentTenantID(ctx)
	if role.TenantId != tenantID {
		return bizerr.NewCode(CodeRoleTenantMismatch)
	}
	return nil
}

// ensureRoleAssignmentUsersMatchRoleBoundary prevents tenant-bound roles from
// being granted outside their tenant/user boundary.
func (s *serviceImpl) ensureRoleAssignmentUsersMatchRoleBoundary(ctx context.Context, role *entity.SysRole, userIDs []int) error {
	if role == nil {
		return bizerr.NewCode(CodeRoleNotFound)
	}
	if len(userIDs) == 0 {
		return nil
	}

	if role.TenantId == datascope.PlatformTenantID {
		count, err := dao.SysUser.Ctx(ctx).
			WhereIn(dao.SysUser.Columns().Id, userIDs).
			Where(do.SysUser{TenantId: datascope.PlatformTenantID}).
			Count()
		if err != nil {
			return err
		}
		if count != len(userIDs) {
			return bizerr.NewCode(CodePlatformRoleAssignmentForbidden)
		}
		return nil
	}

	count, err := dao.SysUser.Ctx(ctx).
		WhereIn(dao.SysUser.Columns().Id, userIDs).
		WhereNot(dao.SysUser.Columns().TenantId, datascope.PlatformTenantID).
		Count()
	if err != nil {
		return err
	}
	if count != len(userIDs) {
		return bizerr.NewCode(CodeTenantRoleAssignmentForbidden)
	}

	if s == nil || s.tenantSvc == nil {
		return nil
	}
	if err := s.tenantSvc.EnsureUsersInTenant(ctx, userIDs, tenantcapsvc.TenantID(role.TenantId)); err != nil {
		if bizerr.Is(err, pkgtenantcap.CodeTenantForbidden) {
			return bizerr.NewCode(CodeTenantRoleAssignmentForbidden)
		}
		return err
	}
	return nil
}

// UnassignUser removes user from a role.
func (s *serviceImpl) UnassignUser(ctx context.Context, roleId int, userId int) error {
	// Check role exists
	_, err := s.GetById(ctx, roleId)
	if err != nil {
		return err
	}
	if err = s.ensureRoleUsersVisible(ctx, []int{userId}); err != nil {
		return err
	}

	urCols := dao.SysUserRole.Columns()
	model := dao.SysUserRole.Ctx(ctx).
		Where(urCols.RoleId, roleId).
		Where(urCols.UserId, userId)
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	_, err = model.Delete()
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
	if err = s.ensureRoleUsersVisible(ctx, userIds); err != nil {
		return err
	}

	urCols := dao.SysUserRole.Columns()
	model := dao.SysUserRole.Ctx(ctx).
		Where(urCols.RoleId, roleId).
		WhereIn(urCols.UserId, userIds)
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	_, err = model.Delete()
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
	m = datascope.ApplyTenantScope(ctx, m, datascope.TenantColumn)
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
	m = datascope.ApplyTenantScope(ctx, m, datascope.TenantColumn)
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
	return s.getUserRoleIdsInScope(ctx, userId)
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
