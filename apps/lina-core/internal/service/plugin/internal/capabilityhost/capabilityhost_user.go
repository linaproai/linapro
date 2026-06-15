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
func (a *userCapabilityAdapter) BatchGet(ctx context.Context, _ capmodel.CapabilityContext, ids []capabilityusercap.UserID) (*capmodel.BatchResult[*capabilityusercap.UserProjection, capabilityusercap.UserID], error) {
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
	model, empty, err = a.applyDataScope(ctx, model)
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

// Search searches visible user candidates with bounded paging.
func (a *userCapabilityAdapter) Search(ctx context.Context, _ capmodel.CapabilityContext, input capabilityusercap.SearchInput) (*capmodel.PageResult[*capabilityusercap.UserProjection], error) {
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
	model, empty, err = a.applyDataScope(ctx, model)
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
func (a *userCapabilityAdapter) applyDataScope(ctx context.Context, model *gdb.Model) (*gdb.Model, bool, error) {
	if a == nil || a.dataScope == nil {
		return model, false, nil
	}
	return a.dataScope.ApplyUserScope(ctx, model, qualifiedSysUserIDColumn())
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
