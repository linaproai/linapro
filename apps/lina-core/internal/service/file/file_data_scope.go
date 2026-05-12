// This file applies shared data-scope rules to file metadata and content access.

package file

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/role"
	"lina-core/pkg/bizerr"
)

// applyFileDataScope constrains file metadata queries by uploader.
func (s *serviceImpl) applyFileDataScope(ctx context.Context, model *gdb.Model) (*gdb.Model, error) {
	model = datascope.ApplyTenantScope(ctx, model, dao.SysFile.Table()+"."+datascope.TenantColumn)
	scopedModel, empty, err := s.currentScopeSvc().ApplyUserScope(ctx, model, qualifiedSysFileCreatedByColumn())
	if err != nil {
		return nil, mapFileDataScopeError(err)
	}
	if empty {
		return model.Where(dao.SysFile.Columns().Id, 0), nil
	}
	return scopedModel, nil
}

// ensureFilesVisible verifies every requested file ID is inside the current data scope.
func (s *serviceImpl) ensureFilesVisible(ctx context.Context, ids []int64) error {
	normalizedIDs := normalizeFileIDs(ids)
	if len(normalizedIDs) == 0 {
		return nil
	}
	model := dao.SysFile.Ctx(ctx).WhereIn(dao.SysFile.Columns().Id, normalizedIDs)
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	err := s.currentScopeSvc().EnsureRowsVisible(ctx, model, qualifiedSysFileCreatedByColumn(), len(normalizedIDs))
	return mapFileDataScopeError(err)
}

// currentScopeSvc returns the shared data-scope service for file operations.
func (s *serviceImpl) currentScopeSvc() datascope.Service {
	return datascope.New(datascope.Dependencies{
		BizCtxSvc: s.bizCtxSvc,
		RoleSvc:   role.Instance(),
		OrgCapSvc: s.orgCapSvc,
	})
}

// mapFileDataScopeError maps shared data-scope errors to file service errors.
func mapFileDataScopeError(err error) error {
	if err == nil {
		return nil
	}
	if bizerr.Is(err, datascope.CodeDataScopeDenied) {
		return bizerr.NewCode(CodeFileDataScopeDenied)
	}
	return err
}

// normalizeFileIDs removes invalid and duplicate file IDs.
func normalizeFileIDs(ids []int64) []int64 {
	result := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// qualifiedSysFileCreatedByColumn returns the fully qualified uploader column.
func qualifiedSysFileCreatedByColumn() string {
	return dao.SysFile.Table() + "." + dao.SysFile.Columns().CreatedBy
}
