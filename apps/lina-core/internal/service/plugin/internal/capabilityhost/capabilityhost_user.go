// This file adapts host user storage to plugin-visible user capability
// contracts without leaking DAO, DO, or entity types.
package capabilityhost

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
)

const (
	statusDisabled = "0"
	statusEnabled  = "1"
)

// Service exposes the user domain service and management commands.
type userCapabilityService interface {
	capabilityusercap.Service
	capabilityusercap.AdminService
}

// adapter implements usercap.Service with database-side tenant scope.
type userCapabilityAdapter struct {
	tenantFilter tenantspi.PluginTableFilterService
	dataScope    datascope.Service
}

var (
	_ capabilityusercap.Service      = (*userCapabilityAdapter)(nil)
	_ capabilityusercap.AdminService = (*userCapabilityAdapter)(nil)
)

// New creates the host-owned user capability adapter.
func newUserCapabilityAdapter(tenantFilter tenantspi.PluginTableFilterService, dataScope datascope.Service) userCapabilityService {
	return &userCapabilityAdapter{tenantFilter: tenantFilter, dataScope: dataScope}
}

// BatchGet returns visible user projections and opaque missing IDs.
func (a *userCapabilityAdapter) BatchGet(ctx context.Context, capCtx capmodel.CapabilityContext, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID], error) {
	result := &capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID]{
		Items:      make(map[capabilityusercap.UserID]*capabilityusercap.UserProjection, len(ids)),
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

	rows := make([]*entity.SysUser, 0, len(parsedIDs))
	cols := dao.SysUser.Columns()
	model := dao.SysUser.Ctx(ctx).
		Fields(cols.Id, cols.TenantId, cols.Username, cols.Nickname, cols.Avatar, cols.Status).
		WhereIn(cols.Id, parsedIDs)
	if a != nil && a.tenantFilter != nil {
		model = a.tenantFilter.Apply(ctx, model, "")
	}
	var (
		empty bool
		err   error
	)
	model, empty, err = a.applyDataScope(ctx, capCtx, model)
	if err != nil {
		return nil, err
	}
	if empty {
		for _, id := range ids {
			if !Contains(result.MissingIDs, id) {
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
		if _, ok := result.Items[id]; !ok && !Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// Current returns the current actor's visible user projection.
func (a *userCapabilityAdapter) Current(ctx context.Context, capCtx capmodel.CapabilityContext) (*capabilityusercap.UserProjection, error) {
	if capCtx.Actor.Type != capmodel.ActorTypeUser || capCtx.Actor.UserID <= 0 {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityActorRequired)
	}
	userID := capabilityusercap.UserID(strconv.FormatInt(capCtx.Actor.UserID, 10))
	result, err := a.BatchGet(ctx, capCtx, []capabilityusercap.UserID{userID})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Items[userID] == nil || len(result.MissingIDs) > 0 {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return result.Items[userID], nil
}

// BatchResolve resolves visible users by IDs, usernames, email addresses, or phone numbers.
func (a *userCapabilityAdapter) BatchResolve(
	ctx context.Context,
	capCtx capmodel.CapabilityContext,
	input capabilityusercap.BatchResolveInput,
) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.ResolveKey], error) {
	result := &capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.ResolveKey]{
		Items:      make(map[capabilityusercap.ResolveKey]*capabilityusercap.UserProjection),
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

	rows := make([]*entity.SysUser, 0, len(resolve.keys))
	cols := dao.SysUser.Columns()
	model := dao.SysUser.Ctx(ctx).
		Fields(cols.Id, cols.TenantId, cols.Username, cols.Nickname, cols.Avatar, cols.Status, cols.Email, cols.Phone)
	model = model.Where(userResolveFilter(model, userResolveColumns{
		id:       cols.Id,
		username: cols.Username,
		email:    cols.Email,
		phone:    cols.Phone,
	}, resolve))
	if a != nil && a.tenantFilter != nil {
		model = a.tenantFilter.Apply(ctx, model, "")
	}
	var (
		empty bool
		err   error
	)
	model, empty, err = a.applyDataScope(ctx, capCtx, model)
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
		if _, ok := result.Items[key]; !ok && !Contains(result.MissingIDs, key) {
			result.MissingIDs = append(result.MissingIDs, key)
		}
	}
	return result, nil
}

// Search searches visible user candidates with bounded paging.
func (a *userCapabilityAdapter) Search(ctx context.Context, capCtx capmodel.CapabilityContext, input capabilityusercap.SearchInput) (*capmodel.PageResult[*capabilityusercap.UserProjection], error) {
	pageNum, pageSize := NormalizePage(input.Page)
	cols := dao.SysUser.Columns()
	model := dao.SysUser.Ctx(ctx)
	if a != nil && a.tenantFilter != nil {
		model = a.tenantFilter.Apply(ctx, model, "")
	}
	var (
		empty bool
		err   error
	)
	model, empty, err = a.applyDataScope(ctx, capCtx, model)
	if err != nil {
		return nil, err
	}
	if empty {
		return &capmodel.PageResult[*capabilityusercap.UserProjection]{Items: []*capabilityusercap.UserProjection{}}, nil
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
		model = model.Where(do.SysUser{Status: 1})
	} else if strings.TrimSpace(string(input.Status)) != "" {
		status, parseErr := strconv.Atoi(strings.TrimSpace(string(input.Status)))
		if parseErr != nil {
			return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		model = model.Where(do.SysUser{Status: status})
	}
	if strings.TrimSpace(string(input.TenantID)) != "" {
		tenantID, parseErr := TenantID(input.TenantID)
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
	items := make([]*capabilityusercap.UserProjection, 0, len(rows))
	for _, row := range rows {
		items = append(items, projectUser(row))
	}
	return &capmodel.PageResult[*capabilityusercap.UserProjection]{Items: items, Total: total}, nil
}

// applyDataScope constrains sys_user reads by the current role data scope.
func (a *userCapabilityAdapter) applyDataScope(ctx context.Context, capCtx capmodel.CapabilityContext, model *gdb.Model) (*gdb.Model, bool, error) {
	if a == nil || a.dataScope == nil {
		return model, false, nil
	}
	if bypassUserDataScope(capCtx) {
		return model, false, nil
	}
	return a.dataScope.ApplyUserScope(ctx, model, qualifiedSysUserIDColumn())
}

// bypassUserDataScope reports whether a host-owned system orchestration call can
// read stable user projections without an HTTP request data-scope snapshot.
func bypassUserDataScope(capCtx capmodel.CapabilityContext) bool {
	return capCtx.SystemCall &&
		capCtx.Actor.Type == capmodel.ActorTypeSystem &&
		capCtx.Source == capmodel.CapabilitySourceHost
}

// EnsureVisible rejects when any requested user is absent or invisible.
func (a *userCapabilityAdapter) EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []capabilityusercap.UserID) error {
	result, err := a.BatchGet(ctx, capCtx, ids)
	if err != nil {
		return err
	}
	if len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// SetStatus changes one visible user's lifecycle status and advances the
// authorization revision after the transaction commits successfully.
func (a *userCapabilityAdapter) SetStatus(ctx context.Context, capCtx capmodel.CapabilityContext, id capabilityusercap.UserID, status string) error {
	normalizedStatus := strings.TrimSpace(status)
	if normalizedStatus != statusDisabled && normalizedStatus != statusEnabled {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	parsedID, err := strconv.Atoi(strings.TrimSpace(string(id)))
	if err != nil || parsedID <= 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if err = a.EnsureVisible(ctx, capCtx, []capabilityusercap.UserID{id}); err != nil {
		return err
	}
	return dao.SysUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		model := tx.Model(dao.SysUser.Table()).Safe().Ctx(ctx).Where(do.SysUser{Id: parsedID})
		if a != nil && a.tenantFilter != nil {
			model = a.tenantFilter.Apply(ctx, model, "")
		}
		result, updateErr := model.Data(do.SysUser{Status: normalizedStatus}).Update()
		if updateErr != nil {
			return updateErr
		}
		affected, affectedErr := result.RowsAffected()
		if affectedErr != nil {
			return affectedErr
		}
		if affected == 0 {
			return bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		return BumpSharedRevision(
			ctx,
			tx,
			AuthorizationCacheDomain,
			AuthorizationCacheScopeGlobal,
			AuthorizationChangeReason,
		)
	})
}

// projectUser converts a host entity into a storage-independent user projection.
func projectUser(row *entity.SysUser) *capabilityusercap.UserProjection {
	if row == nil {
		return nil
	}
	return &capabilityusercap.UserProjection{
		ID:       capabilityusercap.UserID(strconv.Itoa(row.Id)),
		Username: row.Username,
		Nickname: row.Nickname,
		Avatar:   row.Avatar,
		Status:   strconv.Itoa(row.Status),
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
