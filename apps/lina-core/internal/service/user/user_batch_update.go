// This file implements atomic batch updates for selected user-management rows.

package user

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/role"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// BatchUpdateInput defines optional patch fields for selected users.
type BatchUpdateInput struct {
	Ids          []int // User ID list
	UpdateStatus bool  // Whether to update status
	Status       *int  // Status: 1=Normal 0=Disabled
	UpdateRoles  bool  // Whether to replace role assignments
	RoleIds      []int // Role ID list
	UpdateTenant bool  // Whether to replace tenant memberships
	TenantIds    []int // Tenant ID list
}

// BatchUpdate updates selected users atomically.
func (s *serviceImpl) BatchUpdate(ctx context.Context, in BatchUpdateInput) error {
	normalizedIDs := normalizeUserBatchUpdateIDs(in.Ids)
	if len(normalizedIDs) == 0 {
		return bizerr.NewCode(CodeUserBatchUpdateIdsRequired)
	}
	if !in.UpdateStatus && !in.UpdateRoles && !in.UpdateTenant {
		return bizerr.NewCode(CodeUserBatchUpdateFieldsRequired)
	}
	if in.UpdateStatus && in.Status == nil {
		return bizerr.NewCode(CodeUserBatchUpdateStatusRequired)
	}
	if in.UpdateRoles && in.UpdateTenant {
		return bizerr.NewCode(CodeUserBatchUpdateRoleTenantConflict)
	}

	if err := s.ensureBatchUpdateTargetsAllowed(ctx, normalizedIDs, in); err != nil {
		return err
	}

	var (
		tenantPlan *tenantAssignmentBatchPlan
		roleIDs    []int
		err        error
	)
	if in.UpdateTenant {
		tenantPlan, err = s.resolveBatchTenantMemberships(ctx, in.TenantIds)
		if err != nil {
			return err
		}
	}
	if in.UpdateRoles {
		roleIDs, err = s.resolveBatchRoleAssignments(ctx, in.RoleIds, normalizedIDs)
		if err != nil {
			return err
		}
	}

	err = dao.SysUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if in.UpdateStatus {
			updateModel := dao.SysUser.Ctx(ctx).WhereIn(dao.SysUser.Columns().Id, normalizedIDs)
			if _, err := updateModel.Data(do.SysUser{Status: *in.Status}).Update(); err != nil {
				return err
			}
		}

		if in.UpdateTenant && tenantPlan != nil && tenantPlan.shouldReplace {
			for _, userID := range normalizedIDs {
				if _, err := dao.SysUser.Ctx(ctx).
					Where(do.SysUser{Id: userID}).
					Data(do.SysUser{TenantId: tenantPlan.primaryTenantID}).
					Update(); err != nil {
					return err
				}
				if err := s.tenantSvc.ReplaceUserTenantAssignments(ctx, userID, tenantPlan.plan); err != nil {
					return err
				}
			}
		}

		if in.UpdateRoles {
			for _, userID := range normalizedIDs {
				if err := s.replaceUserRoleAssignmentsWithResolvedRoles(ctx, userID, roleIDs); err != nil {
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

// tenantAssignmentBatchPlan carries the host and provider tenant write plan.
type tenantAssignmentBatchPlan struct {
	plan            *pkgtenantcap.UserTenantAssignmentPlan
	shouldReplace   bool
	primaryTenantID int
}

// ensureBatchUpdateTargetsAllowed enforces current-user, visibility, and
// disable-self protections before any mutation is attempted.
func (s *serviceImpl) ensureBatchUpdateTargetsAllowed(ctx context.Context, ids []int, in BatchUpdateInput) error {
	bizCtx := s.bizCtxSvc.Get(ctx)
	for _, id := range ids {
		if bizCtx != nil && bizCtx.UserId == id {
			return bizerr.NewCode(CodeUserCurrentEditDenied)
		}
	}
	if in.UpdateStatus && in.Status != nil && Status(*in.Status) == StatusDisabled {
		for _, id := range ids {
			if bizCtx != nil && bizCtx.UserId == id {
				return bizerr.NewCode(CodeUserCurrentDisableDenied)
			}
		}
	}
	return s.ensureUsersVisible(ctx, ids)
}

// resolveBatchTenantMemberships resolves tenant writes once for all selected users.
func (s *serviceImpl) resolveBatchTenantMemberships(ctx context.Context, tenantIDs []int) (*tenantAssignmentBatchPlan, error) {
	plan, err := s.resolveUpdateTenantMemberships(ctx, tenantIDs)
	if err != nil {
		return nil, err
	}
	return &tenantAssignmentBatchPlan{
		plan:            plan,
		shouldReplace:   plan.ShouldReplace,
		primaryTenantID: int(plan.PrimaryTenant),
	}, nil
}

// resolveBatchRoleAssignments validates role IDs in the active tenant boundary.
func (s *serviceImpl) resolveBatchRoleAssignments(ctx context.Context, roleIDs []int, userIDs []int) ([]int, error) {
	normalizedRoleIDs := normalizeUserBatchUpdateIDs(roleIDs)
	if len(normalizedRoleIDs) == 0 {
		return []int{}, nil
	}
	roles, err := s.visibleRolesByID(ctx, normalizedRoleIDs)
	if err != nil {
		return nil, err
	}
	if len(roles) != len(normalizedRoleIDs) {
		return nil, bizerr.NewCode(role.CodeRoleNotFound)
	}
	for _, item := range roles {
		if item.TenantId != currentTenantID(ctx) {
			return nil, bizerr.NewCode(role.CodeRoleTenantMismatch)
		}
		if err := s.ensureUsersMatchRoleBoundary(ctx, item, userIDs); err != nil {
			return nil, err
		}
	}
	return normalizedRoleIDs, nil
}

// visibleRolesByID returns active roles in the current tenant boundary.
func (s *serviceImpl) visibleRolesByID(ctx context.Context, roleIDs []int) ([]*entity.SysRole, error) {
	var roles []*entity.SysRole
	model := dao.SysRole.Ctx(ctx).
		WhereIn(dao.SysRole.Columns().Id, roleIDs).
		Where(dao.SysRole.Columns().Status, 1)
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	err := model.Scan(&roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// ensureUsersMatchRoleBoundary mirrors role assignment boundaries for host user editing.
func (s *serviceImpl) ensureUsersMatchRoleBoundary(ctx context.Context, item *entity.SysRole, userIDs []int) error {
	if item == nil {
		return bizerr.NewCode(role.CodeRoleNotFound)
	}
	if item.TenantId == datascope.PlatformTenantID {
		count, err := dao.SysUser.Ctx(ctx).
			WhereIn(dao.SysUser.Columns().Id, userIDs).
			Where(do.SysUser{TenantId: datascope.PlatformTenantID}).
			Count()
		if err != nil {
			return err
		}
		if count != len(userIDs) {
			return bizerr.NewCode(role.CodePlatformRoleAssignmentForbidden)
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
		return bizerr.NewCode(role.CodeTenantRoleAssignmentForbidden)
	}
	if s.tenantSvc == nil {
		return nil
	}
	if err := s.tenantSvc.EnsureUsersInTenant(ctx, userIDs, tenantIDFromRole(item)); err != nil {
		return mapTenantMembershipRoleError(err)
	}
	return nil
}

// replaceUserRoleAssignments replaces one user's roles in the active tenant boundary.
func (s *serviceImpl) replaceUserRoleAssignments(ctx context.Context, userID int, roleIDs []int) error {
	normalizedRoleIDs, err := s.resolveBatchRoleAssignments(ctx, roleIDs, []int{userID})
	if err != nil {
		return err
	}
	return s.replaceUserRoleAssignmentsWithResolvedRoles(ctx, userID, normalizedRoleIDs)
}

// replaceUserRoleAssignmentsWithResolvedRoles writes prevalidated role IDs for one user.
func (s *serviceImpl) replaceUserRoleAssignmentsWithResolvedRoles(ctx context.Context, userID int, normalizedRoleIDs []int) error {
	urCols := dao.SysUserRole.Columns()
	deleteModel := dao.SysUserRole.Ctx(ctx).Where(urCols.UserId, userID)
	deleteModel = datascope.ApplyTenantScope(ctx, deleteModel, datascope.TenantColumn)
	if _, err := deleteModel.Delete(); err != nil {
		return err
	}
	if len(normalizedRoleIDs) == 0 {
		return nil
	}
	relations := make([]do.SysUserRole, 0, len(normalizedRoleIDs))
	for _, roleID := range normalizedRoleIDs {
		relations = append(relations, do.SysUserRole{
			UserId:   userID,
			RoleId:   roleID,
			TenantId: currentTenantID(ctx),
		})
	}
	_, err := dao.SysUserRole.Ctx(ctx).Data(relations).Insert()
	return err
}

// tenantIDFromRole converts role ownership to the tenantcap type without
// importing plugin internals.
func tenantIDFromRole(item *entity.SysRole) tenantcapsvc.TenantID {
	if item == nil {
		return tenantcapsvc.TenantID(datascope.PlatformTenantID)
	}
	return tenantcapsvc.TenantID(item.TenantId)
}

// mapTenantMembershipRoleError maps tenantcap visibility errors to role assignment errors.
func mapTenantMembershipRoleError(err error) error {
	if err == nil {
		return nil
	}
	if bizerr.Is(err, pkgtenantcap.CodeTenantForbidden) {
		return bizerr.NewCode(role.CodeTenantRoleAssignmentForbidden)
	}
	return bizerr.NewCode(role.CodeTenantRoleAssignmentForbidden)
}

// normalizeUserBatchUpdateIDs removes invalid and duplicate IDs while preserving request order.
func normalizeUserBatchUpdateIDs(ids []int) []int {
	normalizedIDs := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalizedIDs = append(normalizedIDs, id)
	}
	return normalizedIDs
}
