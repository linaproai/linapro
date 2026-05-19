// This file defines tenant fallback projection metadata for dictionary rows.
// It keeps generated entity structs unchanged while giving management APIs
// stable source and action flags for inherited platform defaults.

package dict

import (
	"context"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
)

// FallbackOverrideMode defines the supported override action mode returned to
// callers for tenant-visible dictionary fallback rows.
type FallbackOverrideMode string

const (
	// FallbackOverrideModeNone means no override action is available.
	FallbackOverrideModeNone FallbackOverrideMode = "none"
	// FallbackOverrideModeCreateTenantOverride means the caller may create a
	// current-tenant row that overrides the inherited platform default.
	FallbackOverrideModeCreateTenantOverride FallbackOverrideMode = "createTenantOverride"
)

// FallbackMetadata describes the source tenant and row-level actions for one
// effective dictionary record.
type FallbackMetadata struct {
	SourceTenantId int
	IsFallback     bool
	CanEdit        bool
	CanOverride    bool
	OverrideMode   FallbackOverrideMode
}

// DictTypeProjection combines one sys_dict_type entity with fallback action
// metadata without modifying generated entity code.
type DictTypeProjection struct {
	*entity.SysDictType
	FallbackMetadata
}

// DictDataProjection combines one sys_dict_data entity with fallback action
// metadata without modifying generated entity code.
type DictDataProjection struct {
	*entity.SysDictData
	FallbackMetadata
}

// ProjectDictType builds fallback action metadata for one dictionary type.
func ProjectDictType(ctx context.Context, record *entity.SysDictType) *DictTypeProjection {
	if record == nil {
		return nil
	}
	return &DictTypeProjection{
		SysDictType:      record,
		FallbackMetadata: dictFallbackMetadata(ctx, record.TenantId, record.AllowTenantOverride),
	}
}

// ProjectDictData builds fallback action metadata for one dictionary data row
// using the supplied platform type override flag.
func ProjectDictData(
	ctx context.Context,
	record *entity.SysDictData,
	allowTenantOverride bool,
) *DictDataProjection {
	if record == nil {
		return nil
	}
	return &DictDataProjection{
		SysDictData:      record,
		FallbackMetadata: dictFallbackMetadata(ctx, record.TenantId, allowTenantOverride),
	}
}

// projectDictTypes builds projections for dictionary type records.
func projectDictTypes(ctx context.Context, records []*entity.SysDictType) []*DictTypeProjection {
	items := make([]*DictTypeProjection, 0, len(records))
	for _, record := range records {
		item := ProjectDictType(ctx, record)
		if item != nil {
			items = append(items, item)
		}
	}
	return items
}

// projectDictData builds projections for dictionary data records. The platform
// type override map is resolved once so large lists do not perform per-row
// dictionary-type queries.
func projectDictData(ctx context.Context, records []*entity.SysDictData) ([]*DictDataProjection, error) {
	overrideAllowedByType, err := dictDataOverridePermissions(ctx, records)
	if err != nil {
		return nil, err
	}
	items := make([]*DictDataProjection, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		item := ProjectDictData(ctx, record, overrideAllowedByType[record.DictType])
		if item != nil {
			items = append(items, item)
		}
	}
	return items, nil
}

// dictFallbackMetadata returns the source and direct-action flags for one
// selected dictionary row.
func dictFallbackMetadata(ctx context.Context, sourceTenantID int, allowTenantOverride bool) FallbackMetadata {
	currentTenantID := datascope.CurrentTenantID(ctx)
	isFallback := currentTenantID != datascope.PlatformTenantID &&
		sourceTenantID == datascope.PlatformTenantID
	canOverride := isFallback && allowTenantOverride
	overrideMode := FallbackOverrideModeNone
	if canOverride {
		overrideMode = FallbackOverrideModeCreateTenantOverride
	}

	return FallbackMetadata{
		SourceTenantId: sourceTenantID,
		IsFallback:     isFallback,
		CanEdit:        !isFallback,
		CanOverride:    canOverride,
		OverrideMode:   overrideMode,
	}
}

// dictDataOverridePermissions returns whether each data row's platform
// dictionary type allows tenant overrides.
func dictDataOverridePermissions(
	ctx context.Context,
	records []*entity.SysDictData,
) (map[string]bool, error) {
	result := make(map[string]bool)
	dictTypes := make([]string, 0, len(records))
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		if record == nil || record.DictType == "" {
			continue
		}
		if _, ok := seen[record.DictType]; ok {
			continue
		}
		seen[record.DictType] = struct{}{}
		dictTypes = append(dictTypes, record.DictType)
	}
	if len(dictTypes) == 0 {
		return result, nil
	}

	var platformTypes []*entity.SysDictType
	err := dao.SysDictType.Ctx(ctx).
		Where(do.SysDictType{TenantId: datascope.PlatformTenantID}).
		WhereIn(dao.SysDictType.Columns().Type, dictTypes).
		Scan(&platformTypes)
	if err != nil {
		return nil, err
	}
	for _, platformType := range platformTypes {
		if platformType != nil {
			result[platformType.Type] = platformType.AllowTenantOverride
		}
	}
	return result, nil
}
