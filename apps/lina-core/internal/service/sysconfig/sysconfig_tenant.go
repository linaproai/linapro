// This file contains tenant fallback helpers for system configuration records.

package sysconfig

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
)

// sysconfigTenantIDs returns the readable tenant ids for configuration
// fallback queries.
func sysconfigTenantIDs(ctx context.Context) []int {
	tenantID := datascope.CurrentTenantID(ctx)
	if tenantID == datascope.PlatformTenantID {
		return []int{datascope.PlatformTenantID}
	}
	return []int{tenantID, datascope.PlatformTenantID}
}

// applySysconfigFallbackScope limits reads to current-tenant and platform
// configuration rows while keeping platform contexts platform-only.
func applySysconfigFallbackScope(ctx context.Context, model *gdb.Model) *gdb.Model {
	return model.WhereIn(datascope.TenantColumn, sysconfigTenantIDs(ctx))
}

// currentTenantConfigDO returns persisted tenant metadata for configuration
// writes.
func currentTenantConfigDO(ctx context.Context) do.SysConfig {
	return do.SysConfig{TenantId: datascope.CurrentTenantID(ctx)}
}

// visibleConfigs collapses fallback query rows into the effective
// configuration view, preferring current-tenant rows over platform rows by key.
func visibleConfigs(ctx context.Context, rows []*entity.SysConfig) []*entity.SysConfig {
	tenantID := datascope.CurrentTenantID(ctx)
	byKey := make(map[string]*entity.SysConfig, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		existing := byKey[row.Key]
		if existing == nil || existing.TenantId != tenantID && row.TenantId == tenantID {
			byKey[row.Key] = row
		}
	}

	result := make([]*entity.SysConfig, 0, len(byKey))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if selected := byKey[row.Key]; selected != nil && selected.Id == row.Id {
			result = append(result, row)
		}
	}
	return result
}

// paginateConfigs returns one page from an already materialized effective
// configuration view.
func paginateConfigs(rows []*entity.SysConfig, pageNum int, pageSize int) []*entity.SysConfig {
	if pageNum <= 0 || pageSize <= 0 {
		return rows
	}
	start := (pageNum - 1) * pageSize
	if start >= len(rows) {
		return []*entity.SysConfig{}
	}
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end]
}
