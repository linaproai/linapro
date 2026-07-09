// This file adapts host user storage to plugin-visible user capability
// contracts without leaking DAO, DO, or entity types.
package capabilityadapter

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	usersvc "lina-core/internal/service/user"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/statusflag"
)

// adapter implements usercap.Service with database-side tenant scope.
type userCapabilityAdapter struct {
	owner        usersvc.Service
	tenantFilter tenantcap.Service
	dataScope    datascope.Service
	bizCtx       bizctxcap.Service
}

var _ capabilityusercap.Service = (*userCapabilityAdapter)(nil)
var _ capabilityusercap.AssignmentService = userAssignmentAdapter{}

// NewCapabilityAdapter creates the host-owned user capability adapter.
func NewCapabilityAdapter(owner usersvc.Service, tenantFilter tenantcap.Service, dataScope datascope.Service, bizCtx bizctxcap.Service) capabilityusercap.Service {
	return &userCapabilityAdapter{owner: owner, tenantFilter: tenantFilter, dataScope: dataScope, bizCtx: bizCtx}
}

// BatchGet returns visible user projections and opaque missing IDs.
func (a *userCapabilityAdapter) BatchGet(ctx context.Context, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID], error) {
	result := &capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.UserID]{
		Items:      make(map[capabilityusercap.UserID]*capabilityusercap.UserInfo, len(ids)),
		MissingIDs: []capabilityusercap.UserID{},
	}
	if len(ids) == 0 {
		return result, nil
	}

	parsedIDs := make([]int, 0, len(ids))
	requested := make(map[int]capabilityusercap.UserID, len(ids))
	for _, id := range ids {
		rawID := strings.TrimSpace(string(id))
		parsedID, err := strconv.Atoi(rawID)
		if err != nil || parsedID <= 0 {
			result.MissingIDs = append(result.MissingIDs, id)
			continue
		}
		if _, exists := requested[parsedID]; exists {
			continue
		}
		requested[parsedID] = id
		parsedIDs = append(parsedIDs, parsedID)
	}
	if len(parsedIDs) == 0 {
		return result, nil
	}

	var (
		rows  = make([]*entity.SysUser, 0, len(parsedIDs))
		cols  = dao.SysUser.Columns()
		model = dao.SysUser.Ctx(ctx).
			Fields(cols.Id, cols.TenantId, cols.Username, cols.Nickname, cols.Avatar, cols.Status).
			WhereIn(cols.Id, parsedIDs)
	)
	if filter := a.pluginTableFilter(); filter != nil {
		model = tenantspi.ApplyPluginTableFilter(ctx, filter, model, "")
	}
	var (
		empty bool
		err   error
	)
	model, empty, err = a.applyDataScope(ctx, model)
	if err != nil {
		return nil, err
	}
	if empty {
		for _, id := range ids {
			if !slices.Contains(result.MissingIDs, id) {
				result.MissingIDs = append(result.MissingIDs, id)
			}
		}
		return result, nil
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		requestID, ok := requested[row.Id]
		if !ok {
			continue
		}
		result.Items[requestID] = projectUser(row)
	}
	for _, id := range ids {
		if _, ok := result.Items[id]; !ok && !slices.Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// Current returns the current actor's visible user projection.
func (a *userCapabilityAdapter) Current(ctx context.Context) (*capabilityusercap.UserInfo, error) {
	current := a.current(ctx)
	if current.UserID <= 0 {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityCurrentUserRequired)
	}
	userID := capabilityusercap.UserID(strconv.Itoa(current.UserID))
	result, err := a.BatchGet(ctx, []capabilityusercap.UserID{userID})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Items[userID] == nil || len(result.MissingIDs) > 0 {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return result.Items[userID], nil
}

// Get returns one visible user projection.
func (a *userCapabilityAdapter) Get(ctx context.Context, id capabilityusercap.UserID) (*capabilityusercap.UserInfo, error) {
	result, err := a.BatchGet(ctx, []capabilityusercap.UserID{id})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if projection := result.Items[id]; projection != nil {
		return projection, nil
	}
	return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
}

// BatchResolve resolves visible users by IDs, usernames, email addresses, or phone numbers.
func (a *userCapabilityAdapter) BatchResolve(
	ctx context.Context,
	input capabilityusercap.BatchResolveInput,
) (*capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.ResolveKey], error) {
	result := &capmodel.BatchResult[*capabilityusercap.UserInfo, capabilityusercap.ResolveKey]{
		Items:      make(map[capabilityusercap.ResolveKey]*capabilityusercap.UserInfo),
		MissingIDs: []capabilityusercap.ResolveKey{},
	}
	if len(input.IDs) > capabilityusercap.MaxBatchResolveIDs ||
		len(input.Usernames) > capabilityusercap.MaxBatchResolveUsernames ||
		len(input.Contacts) > capabilityusercap.MaxBatchResolveContacts {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilityusercap.MaxBatchResolveKeys))
	}

	resolve := normalizeUserResolveInput(input)
	if len(resolve.keys) > capabilityusercap.MaxBatchResolveKeys {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilityusercap.MaxBatchResolveKeys))
	}
	if len(resolve.keys) == 0 {
		return result, nil
	}
	if len(resolve.ids) == 0 && len(resolve.usernames) == 0 && len(resolve.contacts) == 0 {
		result.MissingIDs = append(result.MissingIDs, resolve.keys...)
		return result, nil
	}

	var (
		rows  = make([]*entity.SysUser, 0, len(resolve.keys))
		cols  = dao.SysUser.Columns()
		model = dao.SysUser.Ctx(ctx).
			Fields(cols.Id, cols.TenantId, cols.Username, cols.Nickname, cols.Avatar, cols.Status, cols.Email, cols.Phone)
	)
	model = model.Where(userResolveFilter(model, userResolveColumns{
		id:       cols.Id,
		username: cols.Username,
		email:    cols.Email,
		phone:    cols.Phone,
	}, resolve))
	if filter := a.pluginTableFilter(); filter != nil {
		model = tenantspi.ApplyPluginTableFilter(ctx, filter, model, "")
	}
	var (
		empty bool
		err   error
	)
	model, empty, err = a.applyDataScope(ctx, model)
	if err != nil {
		return nil, err
	}
	if empty {
		result.MissingIDs = append(result.MissingIDs, resolve.keys...)
		return result, nil
	}
	if err = model.OrderAsc(cols.Id).Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		projection := projectUser(row)
		if projection == nil {
			continue
		}
		for _, key := range resolveKeysForUser(row, resolve) {
			if _, exists := result.Items[key]; !exists {
				result.Items[key] = projection
			}
		}
	}
	for _, key := range resolve.keys {
		if _, ok := result.Items[key]; !ok && !slices.Contains(result.MissingIDs, key) {
			result.MissingIDs = append(result.MissingIDs, key)
		}
	}
	return result, nil
}

// List returns visible user candidates with bounded paging.
func (a *userCapabilityAdapter) List(ctx context.Context, input capabilityusercap.ListInput) (*capmodel.PageResult[*capabilityusercap.UserInfo], error) {
	pageNum, pageSize := input.Page.Normalize()
	cols := dao.SysUser.Columns()
	model := dao.SysUser.Ctx(ctx)
	if filter := a.pluginTableFilter(); filter != nil {
		model = tenantspi.ApplyPluginTableFilter(ctx, filter, model, "")
	}
	var (
		empty bool
		err   error
	)
	model, empty, err = a.applyDataScope(ctx, model)
	if err != nil {
		return nil, err
	}
	if empty {
		return &capmodel.PageResult[*capabilityusercap.UserInfo]{Items: []*capabilityusercap.UserInfo{}}, nil
	}
	keyword := strings.TrimSpace(input.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		model = model.Where(
			fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", cols.Username, cols.Nickname),
			like,
			like,
		)
	}
	if input.EnabledOnly {
		model = model.Where(do.SysUser{Status: int(statusflag.EnabledValue)})
	} else if input.Status != nil {
		if *input.Status != statusflag.Disabled && *input.Status != statusflag.EnabledValue {
			return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		model = model.Where(do.SysUser{Status: int(*input.Status)})
	}
	if strings.TrimSpace(string(input.TenantID)) != "" {
		tenantID, parseErr := tenantcap.ParseTenantID(input.TenantID)
		if parseErr != nil || tenantID < 0 {
			return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		model = model.Where(do.SysUser{TenantId: tenantID})
	}
	total, err := model.Clone().Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.SysUser, 0, pageSize)
	if err = model.Clone().
		Fields(cols.Id, cols.TenantId, cols.Username, cols.Nickname, cols.Avatar, cols.Status).
		Page(pageNum, pageSize).
		OrderAsc(cols.Id).
		Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*capabilityusercap.UserInfo, 0, len(rows))
	for _, row := range rows {
		items = append(items, projectUser(row))
	}
	return &capmodel.PageResult[*capabilityusercap.UserInfo]{Items: items, Total: total}, nil
}

// Create creates one governed user through the host user owner.
func (a *userCapabilityAdapter) Create(ctx context.Context, input capabilityusercap.CreateInput) (capabilityusercap.UserID, error) {
	if a == nil || a.owner == nil {
		return "", bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user-owner"))
	}
	status, err := capabilityUserStatus(input.Status)
	if err != nil {
		return "", err
	}
	id, err := a.owner.Create(ctx, usersvc.CreateInput{
		Username:  input.Username,
		Password:  input.Password,
		Nickname:  input.Nickname,
		Email:     input.Email,
		Phone:     input.Phone,
		Sex:       input.Sex,
		Status:    status,
		Remark:    input.Remark,
		DeptId:    input.DeptID,
		PostIds:   input.PostIDs,
		RoleIds:   input.RoleIDs,
		TenantIds: input.TenantIDs,
	})
	if err != nil {
		return "", err
	}
	return capabilityusercap.UserID(strconv.Itoa(id)), nil
}

// ProvisionExternal mints one least-privilege account for a verified external
// identity through the host user owner's system provisioning path. It carries no
// operator context and applies no tenant/role/create-boundary validation; the
// owner derives the username, sets an unusable password, and assigns no roles or
// tenants. Callers own idempotent link de-duplication before invoking it.
func (a *userCapabilityAdapter) ProvisionExternal(ctx context.Context, input capabilityusercap.ProvisionExternalInput) (capabilityusercap.UserID, error) {
	if a == nil || a.owner == nil {
		return "", bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user-owner"))
	}
	id, err := a.owner.ProvisionExternalUser(ctx, usersvc.ProvisionExternalInput{
		Email:          input.Email,
		DisplayName:    input.DisplayName,
		Remark:         input.Remark,
		UsernameAnchor: input.UsernameAnchor,
	})
	if err != nil {
		// Translate the host-internal conflict code into the stable capability
		// sentinel: plugin modules cannot import internal bizerr codes, and the
		// conflict must stay detectable with errors.Is so the provider plugin
		// can apply its caller-visible conflict policy.
		if bizerr.Is(err, usersvc.CodeUserProvisionEmailConflict) {
			return "", capabilityusercap.ErrProvisionEmailConflict
		}
		return "", err
	}
	return capabilityusercap.UserID(strconv.Itoa(id)), nil
}

// Update mutates one visible user through the host user owner.
func (a *userCapabilityAdapter) Update(ctx context.Context, input capabilityusercap.UpdateInput) error {
	if a == nil || a.owner == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user-owner"))
	}
	id, err := parseUserID(input.ID)
	if err != nil {
		return err
	}
	if err = a.EnsureVisible(ctx, []capabilityusercap.UserID{input.ID}); err != nil {
		return err
	}
	return a.owner.Update(ctx, usersvc.UpdateInput{
		Id:        id,
		Username:  input.Username,
		Password:  input.Password,
		Nickname:  input.Nickname,
		Email:     input.Email,
		Phone:     input.Phone,
		Sex:       input.Sex,
		Status:    statusFlagIntPtr(input.Status),
		Remark:    input.Remark,
		DeptId:    input.DeptID,
		PostIds:   input.PostIDs,
		RoleIds:   input.RoleIDs,
		TenantIds: input.TenantIDs,
	})
}

// Delete deletes one visible user through the host user owner.
func (a *userCapabilityAdapter) Delete(ctx context.Context, id capabilityusercap.UserID) error {
	if a == nil || a.owner == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user-owner"))
	}
	parsedID, err := parseUserID(id)
	if err != nil {
		return err
	}
	if err = a.EnsureVisible(ctx, []capabilityusercap.UserID{id}); err != nil {
		return err
	}
	return a.owner.Delete(ctx, parsedID)
}

// ResetPassword resets one visible user's password through the host user owner.
func (a *userCapabilityAdapter) ResetPassword(ctx context.Context, id capabilityusercap.UserID, password string) error {
	if a == nil || a.owner == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user-owner"))
	}
	parsedID, err := parseUserID(id)
	if err != nil {
		return err
	}
	if err = a.EnsureVisible(ctx, []capabilityusercap.UserID{id}); err != nil {
		return err
	}
	return a.owner.ResetPassword(ctx, parsedID, password)
}

// Assignment returns user role assignment operations.
func (a *userCapabilityAdapter) Assignment() capabilityusercap.AssignmentService {
	if a == nil {
		return userAssignmentAdapter{}
	}
	return userAssignmentAdapter{parent: a}
}

// userAssignmentAdapter adapts user-role assignment mutations to the host user owner.
type userAssignmentAdapter struct {
	parent *userCapabilityAdapter
}

// ReplaceRoles replaces one visible user's role assignments through the host user owner.
func (a userAssignmentAdapter) ReplaceRoles(ctx context.Context, id capabilityusercap.UserID, roleIDs []int) error {
	if a.parent == nil || a.parent.owner == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user-owner"))
	}
	parsedID, err := parseUserID(id)
	if err != nil {
		return err
	}
	if err = a.parent.EnsureVisible(ctx, []capabilityusercap.UserID{id}); err != nil {
		return err
	}
	return a.parent.owner.Update(ctx, usersvc.UpdateInput{Id: parsedID, RoleIds: roleIDs})
}

// applyDataScope constrains sys_user reads by the current role data scope.
func (a *userCapabilityAdapter) applyDataScope(ctx context.Context, model *gdb.Model) (*gdb.Model, bool, error) {
	if a == nil || a.dataScope == nil {
		return model, false, nil
	}
	current := a.current(ctx)
	if current.UserID <= 0 && current.PlatformBypass {
		return model, false, nil
	}
	return a.dataScope.ApplyUserScope(ctx, model, qualifiedSysUserIDColumn())
}

func (a *userCapabilityAdapter) current(ctx context.Context) bizctxcap.CurrentContext {
	if a != nil && a.bizCtx != nil {
		return a.bizCtx.Current(ctx)
	}
	return bizctxcap.CurrentFromContext(ctx)
}

func (a *userCapabilityAdapter) pluginTableFilter() tenantcap.FilterService {
	if a == nil || a.tenantFilter == nil {
		return nil
	}
	return a.tenantFilter.Filter()
}

// EnsureVisible rejects when any requested user is absent or invisible.
func (a *userCapabilityAdapter) EnsureVisible(ctx context.Context, ids []capabilityusercap.UserID) error {
	result, err := a.BatchGet(ctx, ids)
	if err != nil {
		return err
	}
	if len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// SetStatus changes one visible user's lifecycle status through the user owner.
func (a *userCapabilityAdapter) SetStatus(ctx context.Context, id capabilityusercap.UserID, status statusflag.Enabled) error {
	if status != statusflag.Disabled && status != statusflag.EnabledValue {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	parsedID, err := parseUserID(id)
	if err != nil {
		return err
	}
	if err = a.EnsureVisible(ctx, []capabilityusercap.UserID{id}); err != nil {
		return err
	}
	if a != nil && a.owner != nil {
		statusValue, statusErr := capabilityUserStatus(&status)
		if statusErr != nil {
			return statusErr
		}
		return a.owner.UpdateStatus(ctx, parsedID, statusValue)
	}
	return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "user-owner"))
}

// capabilityUserStatus converts a plugin-visible user status into the host owner type.
func capabilityUserStatus(status *statusflag.Enabled) (usersvc.Status, error) {
	if status == nil {
		return statusflag.EnabledValue, nil
	}
	switch *status {
	case statusflag.EnabledValue:
		return statusflag.EnabledValue, nil
	case statusflag.Disabled:
		return statusflag.Disabled, nil
	default:
		return 0, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
}

// statusFlagIntPtr converts optional shared status flags to the host owner input type.
func statusFlagIntPtr(status *statusflag.Enabled) *int {
	if status == nil {
		return nil
	}
	value := int(*status)
	return &value
}

// parseUserID decodes one plugin-visible user ID into the host owner key.
func parseUserID(id capabilityusercap.UserID) (int, error) {
	parsedID, err := strconv.Atoi(strings.TrimSpace(string(id)))
	if err != nil || parsedID <= 0 {
		return 0, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return parsedID, nil
}

// projectUser converts a host entity into a storage-independent user projection.
func projectUser(row *entity.SysUser) *capabilityusercap.UserInfo {
	if row == nil {
		return nil
	}
	return &capabilityusercap.UserInfo{
		ID:       capabilityusercap.UserID(strconv.Itoa(row.Id)),
		Username: row.Username,
		Nickname: row.Nickname,
		Avatar:   row.Avatar,
		Status:   statusflag.Enabled(row.Status),
		TenantID: capmodel.DomainID(strconv.Itoa(row.TenantId)),
		Label:    row.Username,
	}
}

// qualifiedSysUserIDColumn returns the user table's fully qualified ID column.
func qualifiedSysUserIDColumn() string {
	return dao.SysUser.Table() + "." + dao.SysUser.Columns().Id
}

type userResolveInput struct {
	keys              []capabilityusercap.ResolveKey
	ids               []int
	usernames         []string
	contacts          []string
	keyByID           map[int]capabilityusercap.ResolveKey
	keyByUsername     map[string]capabilityusercap.ResolveKey
	keyByContact      map[string]capabilityusercap.ResolveKey
	normalizedKeySeen map[capabilityusercap.ResolveKey]struct{}
}

type userResolveColumns struct {
	id       string
	username string
	email    string
	phone    string
}

func normalizeUserResolveInput(input capabilityusercap.BatchResolveInput) userResolveInput {
	out := userResolveInput{
		keys:              []capabilityusercap.ResolveKey{},
		ids:               []int{},
		usernames:         []string{},
		contacts:          []string{},
		keyByID:           map[int]capabilityusercap.ResolveKey{},
		keyByUsername:     map[string]capabilityusercap.ResolveKey{},
		keyByContact:      map[string]capabilityusercap.ResolveKey{},
		normalizedKeySeen: map[capabilityusercap.ResolveKey]struct{}{},
	}
	for _, id := range input.IDs {
		rawID := strings.TrimSpace(string(id))
		key := capabilityusercap.ResolveKey("id:" + rawID)
		out.appendKey(key)
		parsedID, err := strconv.Atoi(rawID)
		if err != nil || parsedID <= 0 {
			continue
		}
		if _, exists := out.keyByID[parsedID]; exists {
			continue
		}
		out.keyByID[parsedID] = key
		out.ids = append(out.ids, parsedID)
	}
	for _, username := range input.Usernames {
		normalizedUsername := strings.TrimSpace(username)
		key := capabilityusercap.ResolveKey("username:" + normalizedUsername)
		out.appendKey(key)
		if normalizedUsername == "" {
			continue
		}
		if _, exists := out.keyByUsername[normalizedUsername]; exists {
			continue
		}
		out.keyByUsername[normalizedUsername] = key
		out.usernames = append(out.usernames, normalizedUsername)
	}
	for _, contact := range input.Contacts {
		normalizedContact := strings.TrimSpace(contact)
		key := capabilityusercap.ResolveKey("contact:" + normalizedContact)
		out.appendKey(key)
		if normalizedContact == "" {
			continue
		}
		if _, exists := out.keyByContact[normalizedContact]; exists {
			continue
		}
		out.keyByContact[normalizedContact] = key
		out.contacts = append(out.contacts, normalizedContact)
	}
	return out
}

func (in *userResolveInput) appendKey(key capabilityusercap.ResolveKey) {
	if _, exists := in.normalizedKeySeen[key]; exists {
		return
	}
	in.normalizedKeySeen[key] = struct{}{}
	in.keys = append(in.keys, key)
}

func userResolveFilter(model *gdb.Model, cols userResolveColumns, resolve userResolveInput) *gdb.WhereBuilder {
	filter := model.Builder()
	hasCondition := false
	if len(resolve.ids) > 0 {
		filter = filter.WhereIn(cols.id, resolve.ids)
		hasCondition = true
	}
	if len(resolve.usernames) > 0 {
		if hasCondition {
			filter = filter.WhereOrIn(cols.username, resolve.usernames)
		} else {
			filter = filter.WhereIn(cols.username, resolve.usernames)
			hasCondition = true
		}
	}
	if len(resolve.contacts) > 0 {
		if hasCondition {
			filter = filter.WhereOrIn(cols.email, resolve.contacts).WhereOrIn(cols.phone, resolve.contacts)
		} else {
			filter = filter.WhereIn(cols.email, resolve.contacts).WhereOrIn(cols.phone, resolve.contacts)
		}
	}
	return filter
}

func resolveKeysForUser(row *entity.SysUser, resolve userResolveInput) []capabilityusercap.ResolveKey {
	if row == nil {
		return nil
	}
	keys := make([]capabilityusercap.ResolveKey, 0, 4)
	if key, ok := resolve.keyByID[row.Id]; ok {
		keys = append(keys, key)
	}
	if key, ok := resolve.keyByUsername[strings.TrimSpace(row.Username)]; ok {
		keys = append(keys, key)
	}
	if key, ok := resolve.keyByContact[strings.TrimSpace(row.Email)]; ok {
		keys = append(keys, key)
	}
	if key, ok := resolve.keyByContact[strings.TrimSpace(row.Phone)]; ok {
		keys = append(keys, key)
	}
	return keys
}
