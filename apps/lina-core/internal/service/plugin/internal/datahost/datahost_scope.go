// This file applies host role data-scope rules to structured data host
// service requests.

package datahost

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
	pkgorgcap "lina-core/pkg/orgcap"
	"lina-core/pkg/pluginbridge"
)

// Host data-scope levels reused by governed data-table access.
const (
	resourceDataScopeNone = 0
	resourceDataScopeAll  = 1
	resourceDataScopeDept = 2
	resourceDataScopeSelf = 3
)

// applyResourceDataScope injects host role data-scope constraints into one data-table query.
func applyResourceDataScope(
	ctx context.Context,
	model *gdb.Model,
	resource *catalog.ResourceSpec,
	identity *pluginbridge.IdentitySnapshotV1,
) (*gdb.Model, error) {
	if model == nil || resource == nil || resource.DataScope == nil {
		return model, nil
	}
	if identity != nil && identity.IsSuperAdmin {
		return model, nil
	}
	if identity == nil || identity.UserID <= 0 {
		return nil, gerror.Newf("data table %s requires user context to apply data scope", resource.Table)
	}

	if identity.DataScopeUnsupported {
		return nil, bizerr.NewCode(
			datascope.CodeDataScopeUnsupported,
			bizerr.P("scope", identity.UnsupportedDataScope),
		)
	}
	switch int(identity.DataScope) {
	case resourceDataScopeAll:
		return model, nil
	case resourceDataScopeDept:
		if resource.DataScope.DeptColumn == "" {
			return model.Where("1 = 0"), nil
		}
		deptIDs, deptErr := getCurrentResourceDeptIDs(ctx, int(identity.UserID))
		if deptErr != nil {
			return nil, deptErr
		}
		if len(deptIDs) == 0 {
			return model.Where("1 = 0"), nil
		}
		return model.WhereIn(resource.DataScope.DeptColumn, deptIDs), nil
	case resourceDataScopeSelf:
		if resource.DataScope.UserColumn == "" {
			return model.Where("1 = 0"), nil
		}
		return model.Where(resource.DataScope.UserColumn, identity.UserID), nil
	default:
		return model.Where("1 = 0"), nil
	}
}

// getCurrentResourceDeptIDs returns the deduplicated department IDs assigned to the user.
func getCurrentResourceDeptIDs(ctx context.Context, userID int) ([]int, error) {
	provider := pkgorgcap.CurrentProvider()
	if provider == nil {
		return []int{}, nil
	}
	return provider.GetUserDeptIDs(ctx, userID)
}
