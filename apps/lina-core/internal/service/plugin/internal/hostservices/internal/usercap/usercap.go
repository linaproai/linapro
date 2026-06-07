// Package usercap adapts host user storage to plugin-visible user capability
// contracts without leaking DAO, DO, or entity types.
package usercap

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/hostservices/internal/domaincap"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
)

const (
	statusDisabled = "0"
	statusEnabled  = "1"
)

// Service exposes the user domain service and management commands.
type Service interface {
	capabilityusercap.Service
	capabilityusercap.AdminService
}

// adapter implements usercap.Service with database-side tenant scope.
type adapter struct {
	tenantFilter tenantcap.PluginTableFilterService
}

var (
	_ capabilityusercap.Service      = (*adapter)(nil)
	_ capabilityusercap.AdminService = (*adapter)(nil)
)

// New creates the host-owned user capability adapter.
func New(tenantFilter tenantcap.PluginTableFilterService) Service {
	return &adapter{tenantFilter: tenantFilter}
}

// BatchGetUsers returns visible user projections and opaque missing IDs.
func (a *adapter) BatchGetUsers(ctx context.Context, _ capmodel.CapabilityContext, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID], error) {
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
		if _, ok := result.Items[id]; !ok && !domaincap.Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// SearchUsers searches visible user candidates with bounded paging.
func (a *adapter) SearchUsers(ctx context.Context, _ capmodel.CapabilityContext, input capabilityusercap.SearchInput) (*capmodel.PageResult[*capabilityusercap.UserProjection], error) {
	pageNum, pageSize := domaincap.NormalizePage(input.Page)
	cols := dao.SysUser.Columns()
	model := dao.SysUser.Ctx(ctx)
	if a != nil && a.tenantFilter != nil {
		model = a.tenantFilter.Apply(ctx, model, "")
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

// EnsureUsersVisible rejects when any requested user is absent or invisible.
func (a *adapter) EnsureUsersVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []capabilityusercap.UserID) error {
	result, err := a.BatchGetUsers(ctx, capCtx, ids)
	if err != nil {
		return err
	}
	if len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// SetUserStatus changes one visible user's lifecycle status and advances the
// authorization revision after the transaction commits successfully.
func (a *adapter) SetUserStatus(ctx context.Context, capCtx capmodel.CapabilityContext, id capabilityusercap.UserID, status string) error {
	normalizedStatus := strings.TrimSpace(status)
	if normalizedStatus != statusDisabled && normalizedStatus != statusEnabled {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	parsedID, err := strconv.Atoi(strings.TrimSpace(string(id)))
	if err != nil || parsedID <= 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if err = a.EnsureUsersVisible(ctx, capCtx, []capabilityusercap.UserID{id}); err != nil {
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
		return domaincap.BumpSharedRevision(
			ctx,
			tx,
			domaincap.AuthorizationCacheDomain,
			domaincap.AuthorizationCacheScopeGlobal,
			domaincap.AuthorizationChangeReason,
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
