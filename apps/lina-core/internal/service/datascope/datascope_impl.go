// datascope_impl.go implements current-user data-scope resolution and query
// filtering. It resolves role scope from the request business context, applies
// organization-provider predicates when available, and returns explicit auth
// or visibility errors instead of allowing callers to infer scope rules.

package datascope

import (
	"context"
	"lina-core/internal/dao"
	"lina-core/pkg/bizerr"

	"github.com/gogf/gf/v2/database/gdb"
)

// Current resolves the current request user's widest enabled role data-scope.
func (s *serviceImpl) Current(ctx context.Context) (*Context, error) {
	if s == nil || s.bizCtxSvc == nil {
		return nil, bizerr.NewCode(CodeDataScopeNotAuthenticated)
	}
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil || bizCtx.UserId <= 0 {
		return nil, bizerr.NewCode(CodeDataScopeNotAuthenticated)
	}

	if s.roleSvc == nil {
		return &Context{UserID: bizCtx.UserId, Scope: ScopeNone}, nil
	}

	snapshot, err := s.roleSvc.GetUserDataScopeSnapshot(ctx, bizCtx.UserId)
	if err != nil {
		return nil, err
	}
	if snapshot == nil || snapshot.UserID != bizCtx.UserId {
		return &Context{UserID: bizCtx.UserId, Scope: ScopeNone}, nil
	}
	return &Context{
		UserID:       snapshot.UserID,
		Scope:        snapshot.Scope,
		IsSuperAdmin: snapshot.IsSuperAdmin,
	}, nil
}

// ApplyUserScope constrains a model by a user-owner column.
func (s *serviceImpl) ApplyUserScope(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error) {
	scopeCtx, err := s.Current(ctx)
	if err != nil {
		return nil, false, err
	}
	return s.applyResolvedScope(ctx, scopeCtx, model, userIDColumn)
}

// ApplyUserScopeWithBypass constrains a model by user scope while preserving
// rows matching a bypass condition, such as built-in scheduled jobs.
func (s *serviceImpl) ApplyUserScopeWithBypass(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	bypassColumn string,
	bypassValue any,
) (*gdb.Model, bool, error) {
	scopeCtx, err := s.Current(ctx)
	if err != nil {
		return nil, false, err
	}
	if scopeCtx.Scope == ScopeAll || scopeCtx.Scope == ScopeTenant {
		return model, false, nil
	}

	builder := model.Builder().Where(bypassColumn, bypassValue)
	switch scopeCtx.Scope {
	case ScopeDept:
		if s.orgCapabilityEnabled(ctx) {
			subQuery, empty, buildErr := s.orgCapSvc.BuildUserDeptScopeExists(ctx, userIDColumn, scopeCtx.UserID)
			if buildErr != nil {
				return nil, false, buildErr
			}
			if !empty {
				builder = builder.WhereOrf("EXISTS ?", subQuery)
			}
			return model.Where(builder), false, nil
		}
		builder = builder.WhereOr(userIDColumn, scopeCtx.UserID)
		return model.Where(builder), false, nil
	case ScopeSelf:
		builder = builder.WhereOr(userIDColumn, scopeCtx.UserID)
		return model.Where(builder), false, nil
	default:
		return model.Where(builder), false, nil
	}
}

// EnsureUsersVisible verifies all target user IDs are visible.
func (s *serviceImpl) EnsureUsersVisible(ctx context.Context, userIDs []int) error {
	normalizedIDs := normalizeUserIDs(userIDs)
	if len(normalizedIDs) == 0 {
		return nil
	}
	model := dao.SysUser.Ctx(ctx).WhereIn(dao.SysUser.Columns().Id, normalizedIDs)
	return s.EnsureRowsVisible(ctx, model, qualifiedColumn(dao.SysUser.Table(), dao.SysUser.Columns().Id), len(normalizedIDs))
}

// EnsureRowsVisible verifies all rows matched by model remain visible after
// scope injection.
func (s *serviceImpl) EnsureRowsVisible(ctx context.Context, model *gdb.Model, userIDColumn string, expectedCount int) error {
	if expectedCount <= 0 {
		return nil
	}
	scopedModel, empty, err := s.ApplyUserScope(ctx, model, userIDColumn)
	if err != nil {
		return err
	}
	if empty {
		return bizerr.NewCode(CodeDataScopeDenied)
	}
	count, err := scopedModel.Count()
	if err != nil {
		return err
	}
	if count != expectedCount {
		return bizerr.NewCode(CodeDataScopeDenied)
	}
	return nil
}

// applyResolvedScope applies one already-resolved scope snapshot to a model.
func (s *serviceImpl) applyResolvedScope(ctx context.Context, scopeCtx *Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error) {
	if scopeCtx == nil {
		return nil, false, bizerr.NewCode(CodeDataScopeNotAuthenticated)
	}
	switch scopeCtx.Scope {
	case ScopeAll, ScopeTenant:
		return model, false, nil
	case ScopeDept:
		if s.orgCapabilityEnabled(ctx) {
			return s.orgCapSvc.ApplyUserDeptScope(ctx, model, userIDColumn, scopeCtx.UserID)
		}
		return model.Where(userIDColumn, scopeCtx.UserID), false, nil
	case ScopeSelf:
		return model.Where(userIDColumn, scopeCtx.UserID), false, nil
	default:
		return model, true, nil
	}
}

// orgCapabilityEnabled reports whether organization capability can participate
// in department-scope filtering.
func (s *serviceImpl) orgCapabilityEnabled(ctx context.Context) bool {
	return s != nil && s.orgCapSvc != nil && s.orgCapSvc.Enabled(ctx)
}

// normalizeUserIDs removes duplicate target IDs for deterministic visibility checks.
func normalizeUserIDs(userIDs []int) []int {
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

// qualifiedColumn returns one fully qualified table column name.
func qualifiedColumn(table string, column string) string {
	return table + "." + column
}
