// This file defines tenant fallback projection metadata for sys_config rows.
// It keeps generated entity structs unchanged while giving API callers stable
// action flags for inherited platform defaults.

package sysconfig

import (
	"context"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
)

// FallbackOverrideMode defines the supported override action mode returned to
// callers for tenant-visible fallback rows.
type FallbackOverrideMode string

const (
	// FallbackOverrideModeNone means no override action is available.
	FallbackOverrideModeNone FallbackOverrideMode = "none"
	// FallbackOverrideModeCreateTenantOverride means the caller may create a
	// current-tenant row that overrides the inherited platform default.
	FallbackOverrideModeCreateTenantOverride FallbackOverrideMode = "createTenantOverride"
)

// FallbackMetadata describes the source tenant and row-level actions for one
// effective configuration record.
type FallbackMetadata struct {
	SourceTenantId int
	IsFallback     bool
	CanEdit        bool
	CanOverride    bool
	OverrideMode   FallbackOverrideMode
}

// ConfigProjection combines one sys_config entity with fallback action
// metadata without modifying generated entity code.
type ConfigProjection struct {
	*entity.SysConfig
	FallbackMetadata
}

// ProjectConfig builds fallback action metadata for one configuration record in
// the supplied request context.
func ProjectConfig(ctx context.Context, record *entity.SysConfig) *ConfigProjection {
	if record == nil {
		return nil
	}
	return &ConfigProjection{
		SysConfig:        record,
		FallbackMetadata: configFallbackMetadata(ctx, record.TenantId),
	}
}

// projectConfigs builds projections for a list of effective configuration
// records.
func projectConfigs(ctx context.Context, records []*entity.SysConfig) []*ConfigProjection {
	items := make([]*ConfigProjection, 0, len(records))
	for _, record := range records {
		item := ProjectConfig(ctx, record)
		if item != nil {
			items = append(items, item)
		}
	}
	return items
}

// configFallbackMetadata returns the source and direct-action flags for one
// selected configuration row.
func configFallbackMetadata(ctx context.Context, sourceTenantID int) FallbackMetadata {
	currentTenantID := datascope.CurrentTenantID(ctx)
	isFallback := currentTenantID != datascope.PlatformTenantID &&
		sourceTenantID == datascope.PlatformTenantID
	canOverride := isFallback
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
