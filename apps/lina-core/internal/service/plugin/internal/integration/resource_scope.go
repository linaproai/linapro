// This file applies host role data-scope rules to plugin-owned generic resource
// queries so dynamic plugins can reuse Lina governance semantics.

package integration

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
	pkgorgcap "lina-core/pkg/orgcap"
)

// Host data-scope levels reused when filtering generic plugin resource queries.
const (
	pluginResourceDataScopeNone = 0
	pluginResourceDataScopeAll  = 1
	pluginResourceDataScopeDept = 2
	pluginResourceDataScopeSelf = 3
)

// applyPluginResourceDataScope injects host role data-scope constraints into one plugin resource query.
func (s *serviceImpl) applyPluginResourceDataScope(
	ctx context.Context,
	model *gdb.Model,
	resource *catalog.ResourceSpec,
) (*gdb.Model, error) {
	if model == nil || resource == nil || resource.DataScope == nil {
		return model, nil
	}

	currentUserID := s.getCurrentPluginResourceUserID(ctx)
	if currentUserID <= 0 {
		return model.Where("1 = 0"), nil
	}

	unsupported, unsupportedValue := s.bizCtxSvc.GetDataScopeUnsupported(ctx)
	if unsupported {
		return nil, bizerr.NewCode(
			datascope.CodeDataScopeUnsupported,
			bizerr.P("scope", unsupportedValue),
		)
	}

	switch s.bizCtxSvc.GetDataScope(ctx) {
	case pluginResourceDataScopeAll:
		return model, nil
	case pluginResourceDataScopeDept:
		if resource.DataScope.DeptColumn == "" {
			return model.Where("1 = 0"), nil
		}
		deptIDs, deptErr := s.getCurrentPluginResourceDeptIDs(ctx, currentUserID)
		if deptErr != nil {
			return nil, deptErr
		}
		if len(deptIDs) == 0 {
			return model.Where("1 = 0"), nil
		}
		return model.WhereIn(resource.DataScope.DeptColumn, deptIDs), nil
	case pluginResourceDataScopeSelf:
		if resource.DataScope.UserColumn == "" {
			return model.Where("1 = 0"), nil
		}
		return model.Where(resource.DataScope.UserColumn, currentUserID), nil
	default:
		return model.Where("1 = 0"), nil
	}
}

// getCurrentPluginResourceUserID returns the current request user ID from the business context.
func (s *serviceImpl) getCurrentPluginResourceUserID(ctx context.Context) int {
	if s.bizCtxSvc == nil {
		return 0
	}
	return s.bizCtxSvc.GetUserId(ctx)
}

// getCurrentPluginResourceDeptIDs returns the deduplicated department IDs for the given user.
func (s *serviceImpl) getCurrentPluginResourceDeptIDs(ctx context.Context, userID int) ([]int, error) {
	provider := pkgorgcap.CurrentProvider()
	if provider == nil {
		return []int{}, nil
	}
	return provider.GetUserDeptIDs(ctx, userID)
}
