// Package membership implements user-to-tenant membership management for the
// multi-tenant source plugin.
package membership

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/bizerr"
	plugincontract "lina-core/pkg/pluginservice/contract"
	pkgtenantcap "lina-core/pkg/tenantcap"
	"lina-plugin-multi-tenant/backend/internal/dao"
	"lina-plugin-multi-tenant/backend/internal/model/do"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// Service defines tenant membership operations.
type Service interface {
	// List queries tenant members by page.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// Add adds a user to one tenant.
	Add(ctx context.Context, in AddInput) (int64, error)
	// Update updates membership status fields.
	Update(ctx context.Context, in UpdateInput) error
	// Remove deletes one membership.
	Remove(ctx context.Context, id int64) error
	// Current returns the current user's membership in the current tenant.
	Current(ctx context.Context) (*Entity, error)
	// GetByUserAndTenant returns one membership for a user and tenant.
	GetByUserAndTenant(ctx context.Context, userID int64, tenantID int64) (*Entity, error)
	// ListUserTenants returns enabled tenant memberships for one user.
	ListUserTenants(ctx context.Context, userID int64) ([]*TenantInfo, error)
	// ApplyUserTenantScope constrains user rows by active current-tenant membership.
	ApplyUserTenantScope(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
	// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
	ApplyUserTenantFilter(ctx context.Context, model *gdb.Model, userIDColumn string, tenantID pkgtenantcap.TenantID) (*gdb.Model, bool, error)
	// ListUserTenantProjections returns tenant ownership labels for visible users.
	ListUserTenantProjections(ctx context.Context, userIDs []int) (map[int]*pkgtenantcap.UserTenantProjection, error)
	// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
	ResolveUserTenantAssignment(ctx context.Context, requested []pkgtenantcap.TenantID, mode pkgtenantcap.UserTenantAssignmentMode) (*pkgtenantcap.UserTenantAssignmentPlan, error)
	// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
	ReplaceUserTenantAssignments(ctx context.Context, userID int, plan *pkgtenantcap.UserTenantAssignmentPlan) error
	// EnsureUsersInTenant verifies every user has active membership in the tenant.
	EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID pkgtenantcap.TenantID) error
	// ValidateStartupConsistency returns user-membership startup consistency failures.
	ValidateStartupConsistency(ctx context.Context) ([]string, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc plugincontract.BizCtxService
}

// New creates and returns a new membership Service instance.
func New(bizCtxSvc plugincontract.BizCtxService) Service {
	return &serviceImpl{bizCtxSvc: bizCtxSvc}
}

// Entity is the service-layer membership projection.
type Entity struct {
	Id       int64  `json:"id" orm:"id"`
	UserID   int64  `json:"userId" orm:"user_id"`
	TenantID int64  `json:"tenantId" orm:"tenant_id"`
	Status   int    `json:"status" orm:"status"`
	JoinedAt string `json:"joinedAt" orm:"joined_at"`
	Username string `json:"username" orm:"username"`
	Nickname string `json:"nickname" orm:"nickname"`
}

// TenantInfo is a user-facing tenant projection.
type TenantInfo struct {
	Id     int64  `json:"id" orm:"id"`
	Code   string `json:"code" orm:"code"`
	Name   string `json:"name" orm:"name"`
	Status string `json:"status" orm:"status"`
}

// userTenantProjectionRow is the joined membership and tenant display
// projection consumed by host user-list screens.
type userTenantProjectionRow struct {
	UserID     int    `json:"userId" orm:"user_id"`
	TenantID   int    `json:"tenantId" orm:"tenant_id"`
	TenantName string `json:"tenantName" orm:"tenant_name"`
}

// membershipTenantRow is the joined membership and tenant lifecycle
// projection used for login and tenant-switch authorization.
type membershipTenantRow struct {
	Id       int64  `json:"id" orm:"id"`
	Status   int    `json:"status" orm:"status"`
	TenantID int64  `json:"tenantId" orm:"tenant_id"`
	TStatus  string `json:"tenantStatus" orm:"tenant_status"`
}

// ListInput defines member list filters.
type ListInput struct {
	PageNum  int
	PageSize int
	TenantID int64
	UserID   int64
	Status   int
}

// ListOutput defines member list output.
type ListOutput struct {
	List  []*Entity
	Total int
}

// AddInput defines membership creation fields.
type AddInput struct {
	TenantID int64
	UserID   int64
}

// UpdateInput defines membership update fields.
type UpdateInput struct {
	Id     int64
	Status *int
}

// sysUserTenantRow is the sys_user tenant projection needed by membership rules.
type sysUserTenantRow struct {
	Id       int64 `json:"id" orm:"id"`
	TenantID int64 `json:"tenantId" orm:"tenant_id"`
}

// List queries tenant members by page.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	in.TenantID = s.effectiveTenantID(ctx, in.TenantID)
	model := membershipListModel(ctx, in)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	list := make([]*Entity, 0)
	if err = membershipListModel(ctx, in).
		Fields("m.*, u.username, u.nickname").
		Page(in.PageNum, in.PageSize).
		OrderDesc("m.id").
		Scan(&list); err != nil {
		return nil, err
	}
	return &ListOutput{List: list, Total: total}, nil
}

// Add adds a user to one tenant.
func (s *serviceImpl) Add(ctx context.Context, in AddInput) (int64, error) {
	in.TenantID = s.effectiveTenantID(ctx, in.TenantID)
	if err := s.ensureUserCanJoinTenant(ctx, in.UserID, []pkgtenantcap.TenantID{pkgtenantcap.TenantID(in.TenantID)}); err != nil {
		return 0, err
	}
	if err := s.ensureTenantAcceptsMembership(ctx, in.TenantID); err != nil {
		return 0, err
	}
	count, err := shared.Model(ctx, shared.TableMembership).
		Where("user_id", in.UserID).
		Where("tenant_id", in.TenantID).
		Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, bizerr.NewCode(CodeMembershipExists)
	}

	bizCtx := s.bizCtxSvc.Current(ctx)
	userID := int64(bizCtx.UserID)
	return shared.Model(ctx, shared.TableMembership).Data(do.UserMembership{
		UserId:    in.UserID,
		TenantId:  in.TenantID,
		Status:    shared.MembershipStatusEnabled,
		CreatedBy: userID,
		UpdatedBy: userID,
	}).InsertAndGetId()
}

// Update updates membership status fields.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	tenantID := s.effectiveTenantID(ctx, shared.PlatformTenantID)
	if _, err := s.getVisible(ctx, in.Id, tenantID); err != nil {
		return err
	}
	bizCtx := s.bizCtxSvc.Current(ctx)
	data := do.UserMembership{UpdatedBy: int64(bizCtx.UserID)}
	if in.Status != nil {
		data.Status = *in.Status
	}
	result, err := s.visibleMembershipModel(ctx, tenantID).Where("id", in.Id).OmitNilData().Data(data).Update()
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return bizerr.NewCode(CodeMembershipNotFound)
	}
	return nil
}

// Remove deletes one membership.
func (s *serviceImpl) Remove(ctx context.Context, id int64) error {
	tenantID := s.effectiveTenantID(ctx, shared.PlatformTenantID)
	if _, err := s.getVisible(ctx, id, tenantID); err != nil {
		return err
	}
	result, err := s.visibleMembershipModel(ctx, tenantID).Where("id", id).Delete()
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return bizerr.NewCode(CodeMembershipNotFound)
	}
	return nil
}

// Current returns the current user's membership in the current tenant.
func (s *serviceImpl) Current(ctx context.Context) (*Entity, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	userID := int64(bizCtx.UserID)
	tenantID := int64(bizCtx.TenantID)
	if userID <= 0 || tenantID <= shared.PlatformTenantID {
		return nil, bizerr.NewCode(CodeMembershipNotFound)
	}
	return s.GetByUserAndTenant(ctx, userID, tenantID)
}

// GetByUserAndTenant returns one membership for a user and tenant.
func (s *serviceImpl) GetByUserAndTenant(ctx context.Context, userID int64, tenantID int64) (*Entity, error) {
	var row *membershipTenantRow
	err := shared.Model(ctx, shared.TableMembership).
		As("m").
		InnerJoin(shared.TableTenant+" t", "t.id = m.tenant_id AND t.deleted_at IS NULL").
		Fields("m.id, m.status, m.tenant_id, t.status AS tenant_status").
		Where("m.user_id", userID).
		Where("m.tenant_id", tenantID).
		Where("m.status", shared.MembershipStatusEnabled).
		Scan(&row)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeMembershipNotFound)
	}
	if row.TStatus != string(shared.TenantStatusActive) {
		return nil, bizerr.NewCode(CodeTenantUnavailable)
	}
	return &Entity{Id: row.Id, UserID: userID, TenantID: row.TenantID, Status: row.Status}, nil
}

// ListUserTenants returns enabled tenant memberships for one user.
func (s *serviceImpl) ListUserTenants(ctx context.Context, userID int64) ([]*TenantInfo, error) {
	list := make([]*TenantInfo, 0)
	err := shared.Model(ctx, shared.TableMembership).As("m").
		InnerJoin(shared.TableTenant+" t", "t.id = m.tenant_id").
		Fields("t.id, t.code, t.name, t.status").
		Where("m.user_id", userID).
		Where("m.status", shared.MembershipStatusEnabled).
		Where("t.status", string(shared.TenantStatusActive)).
		OrderAsc("m.id").
		Scan(&list)
	return list, err
}

// ApplyUserTenantScope constrains user rows by active current-tenant membership.
func (s *serviceImpl) ApplyUserTenantScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
) (*gdb.Model, bool, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	tenantID := int64(bizCtx.TenantID)
	if model == nil || tenantID <= shared.PlatformTenantID {
		return model, false, nil
	}
	return model.WhereIn(userIDColumn, activeMembershipUserModel(ctx, tenantID)), false, nil
}

// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
func (s *serviceImpl) ApplyUserTenantFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	tenantID pkgtenantcap.TenantID,
) (*gdb.Model, bool, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	if model == nil ||
		tenantID <= pkgtenantcap.PLATFORM ||
		bizCtx.TenantID != int(shared.PlatformTenantID) {
		return model, false, nil
	}
	return model.WhereIn(userIDColumn, activeMembershipUserModel(ctx, int64(tenantID))), false, nil
}

// ListUserTenantProjections returns tenant ownership labels for visible users.
func (s *serviceImpl) ListUserTenantProjections(
	ctx context.Context,
	userIDs []int,
) (map[int]*pkgtenantcap.UserTenantProjection, error) {
	result := make(map[int]*pkgtenantcap.UserTenantProjection)
	if len(userIDs) == 0 {
		return result, nil
	}

	var rows []*userTenantProjectionRow
	model := shared.Model(ctx, shared.TableMembership).As("m").
		InnerJoin(shared.TableTenant+" t", "t.id = m.tenant_id AND t.deleted_at IS NULL").
		Fields("m.user_id, m.tenant_id, t.name AS tenant_name").
		WhereIn("m.user_id", userIDs).
		Where("m.status", shared.MembershipStatusEnabled).
		Where("t.status", string(shared.TenantStatusActive)).
		OrderAsc("m.id")
	bizCtx := s.bizCtxSvc.Current(ctx)
	if tenantID := int64(bizCtx.TenantID); tenantID > shared.PlatformTenantID {
		model = model.Where("m.tenant_id", tenantID)
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		item := result[row.UserID]
		if item == nil {
			item = &pkgtenantcap.UserTenantProjection{}
			result[row.UserID] = item
		}
		item.TenantIDs = append(item.TenantIDs, pkgtenantcap.TenantID(row.TenantID))
		item.TenantNames = append(item.TenantNames, row.TenantName)
	}
	return result, nil
}

// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
func (s *serviceImpl) ResolveUserTenantAssignment(
	ctx context.Context,
	requested []pkgtenantcap.TenantID,
	mode pkgtenantcap.UserTenantAssignmentMode,
) (*pkgtenantcap.UserTenantAssignmentPlan, error) {
	normalized := normalizeTenantIDs(requested)
	bizCtx := s.bizCtxSvc.Current(ctx)
	currentTenantID := int64(bizCtx.TenantID)
	if currentTenantID > shared.PlatformTenantID {
		if mode == pkgtenantcap.UserTenantAssignmentUpdate {
			return &pkgtenantcap.UserTenantAssignmentPlan{}, nil
		}
		return &pkgtenantcap.UserTenantAssignmentPlan{
			TenantIDs:     []pkgtenantcap.TenantID{pkgtenantcap.TenantID(currentTenantID)},
			ShouldReplace: true,
			PrimaryTenant: pkgtenantcap.TenantID(currentTenantID),
		}, nil
	}
	if err := s.ensureTenantIDsAcceptMembership(ctx, normalized); err != nil {
		return nil, err
	}
	return &pkgtenantcap.UserTenantAssignmentPlan{
		TenantIDs:     normalized,
		ShouldReplace: mode == pkgtenantcap.UserTenantAssignmentUpdate || len(normalized) > 0,
		PrimaryTenant: firstTenantIDOrPlatform(normalized),
	}, nil
}

// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
func (s *serviceImpl) ReplaceUserTenantAssignments(
	ctx context.Context,
	userID int,
	plan *pkgtenantcap.UserTenantAssignmentPlan,
) error {
	if userID <= 0 || plan == nil {
		return nil
	}
	normalized := normalizeTenantIDs(plan.TenantIDs)
	if len(normalized) > 0 {
		if err := s.ensureUserCanJoinTenant(ctx, int64(userID), normalized); err != nil {
			return err
		}
	}
	if len(normalized) == 0 {
		if _, err := dao.UserMembership.Ctx(ctx).Unscoped().Where("user_id", userID).Delete(); err != nil {
			return err
		}
		return nil
	}
	if err := s.ensureTenantIDsAcceptMembership(ctx, normalized); err != nil {
		return err
	}

	if _, err := dao.UserMembership.Ctx(ctx).Unscoped().Where("user_id", userID).Delete(); err != nil {
		return err
	}
	bizCtx := s.bizCtxSvc.Current(ctx)
	operatorUserID := int64(bizCtx.UserID)
	for _, tenantID := range normalized {
		if _, err := dao.UserMembership.Ctx(ctx).Data(do.UserMembership{
			UserId:    userID,
			TenantId:  int64(tenantID),
			Status:    shared.MembershipStatusEnabled,
			CreatedBy: operatorUserID,
			UpdatedBy: operatorUserID,
		}).Insert(); err != nil {
			return err
		}
	}
	return nil
}

// EnsureUsersInTenant verifies every user has active membership in the tenant.
func (s *serviceImpl) EnsureUsersInTenant(
	ctx context.Context,
	userIDs []int,
	tenantID pkgtenantcap.TenantID,
) error {
	normalized := normalizeUserIDs(userIDs)
	if len(normalized) == 0 || tenantID <= pkgtenantcap.PLATFORM {
		return nil
	}
	count, err := activeMembershipUserModel(ctx, int64(tenantID)).WhereIn("user_id", normalized).Count()
	if err != nil {
		return err
	}
	if count != len(normalized) {
		return bizerr.NewCode(pkgtenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	return nil
}

// ValidateStartupConsistency returns user-membership startup consistency failures.
func (s *serviceImpl) ValidateStartupConsistency(ctx context.Context) ([]string, error) {
	rows := make([]struct {
		Id       int    `json:"id" orm:"id"`
		Username string `json:"username" orm:"username"`
	}, 0)
	err := shared.Model(ctx, shared.TableSysUser).
		As("u").
		Fields("u.id, u.username").
		InnerJoin(
			shared.TableMembership+" m",
			"m.user_id = u.id AND m.deleted_at IS NULL AND m.status = 1",
		).
		Where("u.tenant_id", shared.PlatformTenantID).
		Limit(10).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	details := make([]string, 0, len(rows))
	for _, row := range rows {
		details = append(details, "platform user "+row.Username+"("+gconv.String(row.Id)+") must not have active tenant membership")
	}
	return details, nil
}

// getVisible retrieves one membership by primary key within the visible tenant.
func (s *serviceImpl) getVisible(ctx context.Context, id int64, tenantID int64) (*Entity, error) {
	var item *Entity
	if err := s.visibleMembershipModel(ctx, tenantID).Where("id", id).Scan(&item); err != nil {
		return nil, err
	}
	if item == nil {
		return nil, bizerr.NewCode(CodeMembershipNotFound)
	}
	return item, nil
}

// ensureUserCanJoinTenant enforces platform-user and cardinality membership rules.
func (s *serviceImpl) ensureUserCanJoinTenant(
	ctx context.Context,
	userID int64,
	replacementTenantIDs []pkgtenantcap.TenantID,
) error {
	var user *sysUserTenantRow
	if err := shared.Model(ctx, shared.TableSysUser).Fields("id", "tenant_id").Where("id", userID).Scan(&user); err != nil {
		return err
	}
	if user != nil && user.TenantID == shared.PlatformTenantID && len(replacementTenantIDs) > 0 {
		return bizerr.NewCode(CodePlatformMembershipForbidden)
	}
	if membershipCardinalityMode() != shared.SingleCardinality {
		return nil
	}
	if len(replacementTenantIDs) > 1 {
		return bizerr.NewCode(CodeSingleCardinalityViolation)
	}
	return nil
}

// membershipCardinalityMode returns the code-owned tenant cardinality default.
// The default is multi, meaning one global user identity may belong to multiple
// tenants. A future management setting can switch this to single without
// reintroducing host config-file coupling.
func membershipCardinalityMode() string {
	return shared.DefaultCardinality
}

// ensureTenantAcceptsMembership verifies membership writes target an active tenant.
func (s *serviceImpl) ensureTenantAcceptsMembership(ctx context.Context, tenantID int64) error {
	if tenantID <= shared.PlatformTenantID {
		return bizerr.NewCode(CodeTenantUnavailable)
	}
	var tenant *struct {
		Status string `json:"status" orm:"status"`
	}
	if err := shared.Model(ctx, shared.TableTenant).Fields("status").Where("id", tenantID).Scan(&tenant); err != nil {
		return err
	}
	if tenant == nil || tenant.Status != string(shared.TenantStatusActive) {
		return bizerr.NewCode(CodeTenantUnavailable)
	}
	return nil
}

// ensureTenantIDsAcceptMembership verifies all requested tenant IDs are active.
func (s *serviceImpl) ensureTenantIDsAcceptMembership(ctx context.Context, tenantIDs []pkgtenantcap.TenantID) error {
	for _, tenantID := range tenantIDs {
		if err := s.ensureTenantAcceptsMembership(ctx, int64(tenantID)); err != nil {
			return err
		}
	}
	return nil
}

// effectiveTenantID returns the current request tenant as the authority for
// tenant-scoped member operations. Platform requests may still specify a tenant
// explicitly for administration and E2E setup.
func (s *serviceImpl) effectiveTenantID(ctx context.Context, requestedTenantID int64) int64 {
	currentTenantID := int64(s.bizCtxSvc.Current(ctx).TenantID)
	if currentTenantID > shared.PlatformTenantID {
		return currentTenantID
	}
	return requestedTenantID
}

// visibleMembershipModel builds a membership model constrained by tenant when
// the current request is tenant-scoped.
func (s *serviceImpl) visibleMembershipModel(ctx context.Context, tenantID int64) *gdb.Model {
	model := shared.Model(ctx, shared.TableMembership)
	if tenantID > shared.PlatformTenantID {
		model = model.Where("tenant_id", tenantID)
	}
	return model
}

// membershipListModel builds the shared member list query without a projection
// so Count can generate valid SQL.
func membershipListModel(ctx context.Context, in ListInput) *gdb.Model {
	model := shared.Model(ctx, shared.TableMembership).As("m").
		LeftJoin(shared.TableSysUser+" u", "u.id = m.user_id")
	if in.TenantID > 0 {
		model = model.Where("m.tenant_id", in.TenantID)
	}
	if in.UserID > 0 {
		model = model.Where("m.user_id", in.UserID)
	}
	if in.Status >= 0 {
		model = model.Where("m.status", in.Status)
	}
	return model
}

// activeMembershipUserModel returns the subquery for active users in one tenant.
func activeMembershipUserModel(ctx context.Context, tenantID int64) *gdb.Model {
	return dao.UserMembership.Ctx(ctx).
		Fields("user_id").
		Where("tenant_id", tenantID).
		Where("status", shared.MembershipStatusEnabled)
}

// normalizeTenantIDs returns positive unique tenant IDs while preserving order.
func normalizeTenantIDs(tenantIDs []pkgtenantcap.TenantID) []pkgtenantcap.TenantID {
	normalized := make([]pkgtenantcap.TenantID, 0, len(tenantIDs))
	seen := make(map[pkgtenantcap.TenantID]struct{}, len(tenantIDs))
	for _, tenantID := range tenantIDs {
		if tenantID <= pkgtenantcap.PLATFORM {
			continue
		}
		if _, ok := seen[tenantID]; ok {
			continue
		}
		seen[tenantID] = struct{}{}
		normalized = append(normalized, tenantID)
	}
	return normalized
}

// firstTenantIDOrPlatform returns the primary tenant for host sys_user writes.
func firstTenantIDOrPlatform(tenantIDs []pkgtenantcap.TenantID) pkgtenantcap.TenantID {
	if len(tenantIDs) == 0 {
		return pkgtenantcap.PLATFORM
	}
	return tenantIDs[0]
}

// normalizeUserIDs removes invalid and duplicate user IDs while preserving order.
func normalizeUserIDs(userIDs []int) []int {
	normalized := make([]int, 0, len(userIDs))
	seen := make(map[int]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID <= 0 {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		normalized = append(normalized, userID)
	}
	return normalized
}
