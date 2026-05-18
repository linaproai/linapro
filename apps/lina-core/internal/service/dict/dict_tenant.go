// This file contains tenant fallback and override guard helpers for dictionary
// records.

package dict

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
)

// dictTenantIDs returns the readable tenant ids for dictionary fallback
// queries. Tenant contexts see their own rows first and platform defaults
// second; platform contexts keep the platform-only view.
func dictTenantIDs(ctx context.Context) []int {
	tenantID := datascope.CurrentTenantID(ctx)
	if tenantID == datascope.PlatformTenantID {
		return []int{datascope.PlatformTenantID}
	}
	return []int{tenantID, datascope.PlatformTenantID}
}

// applyDictFallbackScope limits dictionary reads to current-tenant and
// platform rows when tenant fallback is active.
func applyDictFallbackScope(ctx context.Context, model *gdb.Model) *gdb.Model {
	return model.WhereIn(datascope.TenantColumn, dictTenantIDs(ctx))
}

// currentTenantDictDO returns persisted tenant metadata for dictionary writes.
func currentTenantDictDO(ctx context.Context) do.SysDictType {
	return do.SysDictType{TenantId: datascope.CurrentTenantID(ctx)}
}

// currentTenantDictDataDO returns persisted tenant metadata for dictionary data
// writes.
func currentTenantDictDataDO(ctx context.Context) do.SysDictData {
	return do.SysDictData{TenantId: datascope.CurrentTenantID(ctx)}
}

// visibleDictTypes collapses fallback query rows into the effective dictionary
// type view, preferring current-tenant rows over platform rows by type.
func visibleDictTypes(ctx context.Context, rows []*entity.SysDictType) []*entity.SysDictType {
	tenantID := datascope.CurrentTenantID(ctx)
	byType := make(map[string]*entity.SysDictType, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		existing := byType[row.Type]
		if existing == nil || existing.TenantId != tenantID && row.TenantId == tenantID {
			byType[row.Type] = row
		}
	}

	result := make([]*entity.SysDictType, 0, len(byType))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if selected := byType[row.Type]; selected != nil && selected.Id == row.Id {
			result = append(result, row)
		}
	}
	return result
}

// visibleDictData collapses fallback query rows into the effective dictionary
// data view, preferring current-tenant rows over platform rows by type/value.
func visibleDictData(ctx context.Context, rows []*entity.SysDictData) []*entity.SysDictData {
	tenantID := datascope.CurrentTenantID(ctx)
	byValue := make(map[string]*entity.SysDictData, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		key := row.DictType + "\x00" + row.Value
		existing := byValue[key]
		if existing == nil || existing.TenantId != tenantID && row.TenantId == tenantID {
			byValue[key] = row
		}
	}

	result := make([]*entity.SysDictData, 0, len(byValue))
	for _, row := range rows {
		if row == nil {
			continue
		}
		key := row.DictType + "\x00" + row.Value
		if selected := byValue[key]; selected != nil && selected.Id == row.Id {
			result = append(result, row)
		}
	}
	return result
}

// assertDictTenantOverrideAllowed rejects tenant dictionary overrides unless
// the platform dictionary type explicitly allows them.
func assertDictTenantOverrideAllowed(ctx context.Context, dictType string) error {
	tenantID := datascope.CurrentTenantID(ctx)
	if tenantID == datascope.PlatformTenantID {
		return nil
	}

	var platformType *entity.SysDictType
	err := dao.SysDictType.Ctx(ctx).
		Where(do.SysDictType{
			TenantId: datascope.PlatformTenantID,
			Type:     dictType,
		}).
		Scan(&platformType)
	if err != nil {
		return err
	}
	if platformType == nil {
		return nil
	}
	if !platformType.AllowTenantOverride {
		return bizerr.NewCode(CodeDictTypeTenantOverrideDenied, bizerr.P("type", dictType))
	}
	return nil
}
