// Package sysconfig implements system-configuration query, mutation, import,
// and export services for the Lina core host service.
package sysconfig

import (
	"context"
	"io"

	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/pkg/configvaluetype"
)

// Service defines the sysconfig service contract.
type Service interface {
	// List queries tenant-visible config records with pagination and optional
	// name, key, and creation-time filters. Data is filtered by tenant fallback
	// scope before pagination and labels are localized when an i18n translator
	// is configured. Database errors are returned unchanged.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById retrieves one config record by ID within the current tenant data
	// scope for edit/detail display. Name and remark are localized for the
	// request language when an i18n translator is configured; value remains the
	// raw stored text so editable values are not overwritten by display
	// projections. Missing or out-of-scope records return a sysconfig not-found
	// business error.
	GetById(ctx context.Context, id int64) (*entity.SysConfig, error)
	// Create creates a new config record in the current tenant scope. Protected
	// runtime/public frontend keys are validated through the host config service,
	// duplicate keys return business errors, and sys_config runtime snapshots
	// are refreshed after successful creation.
	Create(ctx context.Context, in CreateInput) (int64, error)
	// Update updates an existing config record in the current tenant scope.
	// Built-in protected keys cannot be renamed, built-in name and remark are
	// ignored on write so localized edit projections cannot pollute storage,
	// duplicate keys are rejected, protected values are validated, and
	// sys_config runtime snapshots are refreshed after effective key or value
	// changes.
	Update(ctx context.Context, in UpdateInput) error
	// Delete soft-deletes a config record using GoFrame's auto soft-delete
	// feature after tenant-scope visibility and built-in protection checks.
	Delete(ctx context.Context, id int64) error
	// GetByKey retrieves one tenant-specific or fallback platform config by key.
	// Missing keys return a sysconfig key-not-found business error and returned
	// records are localized and decorated with fallback metadata when i18n is
	// available.
	GetByKey(ctx context.Context, key string) (*ConfigProjection, error)
	// Export generates an Excel file with tenant-visible config data. Optional
	// filters or explicit IDs are applied before visibility filtering; Excel
	// write errors and database errors are returned.
	Export(ctx context.Context, in ExportInput) (data []byte, err error)
	// Import reads an Excel file and creates configs from it.
	// If updateSupport is true, existing tenant-visible records matched by key
	// are updated; otherwise, they are skipped. Protected values and runtime
	// snapshot refresh constraints match Create and Update.
	Import(ctx context.Context, fileReader io.Reader, updateSupport bool) (result *ImportResult, err error)
	// GenerateImportTemplate creates a localized Excel template for config
	// import. The template has no side effects and returns Excel generation
	// errors to the caller.
	GenerateImportTemplate(ctx context.Context) (data []byte, err error)
}

// Interface compliance assertion for the default sysconfig service
// implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	configSvc hostconfig.Service
	i18nSvc   i18nsvc.Service
}

// New creates a sysconfig service from explicit runtime-owned dependencies.
func New(configSvc hostconfig.Service, i18nSvc i18nsvc.Service) Service {
	return &serviceImpl{
		configSvc: configSvc,
		i18nSvc:   i18nSvc,
	}
}

// ListInput defines the sysconfig list query input.
type ListInput struct {
	PageNum   int
	PageSize  int
	Name      string
	Key       string
	BeginTime string
	EndTime   string
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*ConfigProjection // Config list
	Total int                 // Total count
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Name      string                   // Parameter name
	Key       string                   // Parameter key
	Value     string                   // Parameter value
	ValueType string                   // Parameter value input type; empty defaults to text
	Options   []configvaluetype.Option // Selectable options for enum-like types
	Remark    string                   // Remark
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id        int64                     // Parameter ID
	Name      *string                   // Parameter name
	Key       *string                   // Parameter key
	Value     *string                   // Parameter value
	ValueType *string                   // Parameter value input type
	Options   *[]configvaluetype.Option // Selectable options for enum-like types
	Remark    *string                   // Remark
}

// ExportInput defines input for Export function.
type ExportInput struct {
	Name      string  // Parameter name, supports fuzzy search
	Key       string  // Parameter key, supports fuzzy search
	BeginTime string  // Creation time start
	EndTime   string  // Creation time end
	Ids       []int64 // Specific IDs to export; if empty, export all matching records
}
