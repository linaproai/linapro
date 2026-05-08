// Package datascope implements shared role data-permission resolution and
// database-scope injection for host-owned resources.
package datascope

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/orgcap"
	"lina-core/pkg/bizerr"
)

// Scope represents the effective data range stored on enabled roles.
type Scope int

// Role data-scope values follow sys_role.data_scope.
const (
	// ScopeNone denies governed resource access.
	ScopeNone Scope = 0
	// ScopeAll grants access to all governed rows.
	ScopeAll Scope = 1
	// ScopeDept grants access to rows owned by users in the current department scope.
	ScopeDept Scope = 2
	// ScopeSelf grants access only to rows owned by the current user.
	ScopeSelf Scope = 3
)

// AccessSnapshot stores the effective role-governed data scope for one user.
type AccessSnapshot struct {
	UserID       int   // UserID owns the effective data-scope snapshot.
	Scope        Scope // Scope is the widest enabled role data-scope for the user.
	IsSuperAdmin bool  // IsSuperAdmin reports whether the user bypasses role data-scope checks.
}

// AccessProvider is the narrow role dependency needed to resolve cached data scopes.
type AccessProvider interface {
	// GetUserDataScopeSnapshot returns the user's effective role data-scope snapshot.
	GetUserDataScopeSnapshot(ctx context.Context, userID int) (*AccessSnapshot, error)
}

// Dependencies groups optional collaborators for the data-scope service.
type Dependencies struct {
	BizCtxSvc bizctx.Service // BizCtxSvc resolves the current authenticated user.
	RoleSvc   AccessProvider // RoleSvc resolves effective role data-scope snapshots.
	OrgCapSvc orgcap.Service // OrgCapSvc applies optional organization-aware constraints.
}

// Context stores the resolved data-permission snapshot for one request.
type Context struct {
	UserID       int   // UserID is the authenticated operator user ID.
	Scope        Scope // Scope is the widest effective data-scope.
	IsSuperAdmin bool  // IsSuperAdmin reports whether the user bypasses data scope.
}

// Service defines shared data-scope operations used by host modules.
type Service interface {
	// Current resolves the current request user's effective data-scope snapshot.
	Current(ctx context.Context) (*Context, error)
	// ApplyUserScope constrains a model by a user-owner column.
	ApplyUserScope(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
	// ApplyUserScopeWithBypass constrains a model by a user-owner column while
	// preserving rows that match an explicit bypass condition.
	ApplyUserScopeWithBypass(ctx context.Context, model *gdb.Model, userIDColumn string, bypassColumn string, bypassValue any) (*gdb.Model, bool, error)
	// EnsureUsersVisible verifies all target user IDs are visible.
	EnsureUsersVisible(ctx context.Context, userIDs []int) error
	// EnsureRowsVisible verifies all rows matched by model remain visible after scope injection.
	EnsureRowsVisible(ctx context.Context, model *gdb.Model, userIDColumn string, expectedCount int) error
}

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc bizctx.Service
	roleSvc   AccessProvider
	orgCapSvc orgcap.Service
}

// New creates one shared data-scope service.
func New(deps Dependencies) Service {
	if deps.BizCtxSvc == nil {
		deps.BizCtxSvc = bizctx.New()
	}
	if deps.OrgCapSvc == nil {
		deps.OrgCapSvc = orgcap.New(nil)
	}
	return &serviceImpl{
		bizCtxSvc: deps.BizCtxSvc,
		roleSvc:   deps.RoleSvc,
		orgCapSvc: deps.OrgCapSvc,
	}
}

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
	if scopeCtx.Scope == ScopeAll {
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
	case ScopeAll:
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
