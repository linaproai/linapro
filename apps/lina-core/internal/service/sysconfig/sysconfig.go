// Package sysconfig implements system-configuration query, mutation, import,
// and export services for the Lina core host service.
package sysconfig

import (
	"context"
	"io"

	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
)

// Service defines the sysconfig service contract.
type Service interface {
	// List queries tenant-visible config records with pagination and optional
	// name, key, and creation-time filters. Data is filtered by tenant fallback
	// scope before pagination and labels are localized when an i18n translator
	// is configured. Database errors are returned unchanged.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById retrieves one config record by ID within the current tenant data
	// scope. Missing or out-of-scope records return a sysconfig not-found
	// business error.
	GetById(ctx context.Context, id int) (*entity.SysConfig, error)
	// Create creates a new config record in the current tenant scope. Protected
	// runtime/public frontend keys are validated through the host config service,
	// duplicate keys return business errors, and runtime-parameter cache
	// snapshots are refreshed when affected.
	Create(ctx context.Context, in CreateInput) (int, error)
	// Update updates an existing config record in the current tenant scope.
	// Built-in protected keys cannot be renamed, duplicate keys are rejected,
	// protected values are validated, and runtime-parameter cache snapshots are
	// refreshed when affected.
	Update(ctx context.Context, in UpdateInput) error
	// Delete soft-deletes a config record using GoFrame's auto soft-delete
	// feature after tenant-scope visibility and built-in protection checks.
	Delete(ctx context.Context, id int) error
	// GetByKey retrieves one tenant-specific or fallback platform config by key.
	// Missing keys return a sysconfig key-not-found business error and returned
	// records are localized when i18n is available.
	GetByKey(ctx context.Context, key string) (*entity.SysConfig, error)
	// Export generates an Excel file with tenant-visible config data. Optional
	// filters or explicit IDs are applied before visibility filtering; Excel
	// write errors and database errors are returned.
	Export(ctx context.Context, in ExportInput) (data []byte, err error)
	// Import reads an Excel file and creates configs from it.
	// If updateSupport is true, existing tenant-visible records matched by key
	// are updated; otherwise, they are skipped. Protected values and runtime
	// parameter refresh constraints match Create and Update.
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
	i18nSvc   sysconfigI18nTranslator
}

// New creates a sysconfig service from explicit runtime-owned dependencies.
func New(configSvc hostconfig.Service, i18nSvc sysconfigI18nTranslator) Service {
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
	List  []*entity.SysConfig // Config list
	Total int                 // Total count
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Name   string // Parameter name
	Key    string // Parameter key
	Value  string // Parameter value
	Remark string // Remark
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id     int     // Parameter ID
	Name   *string // Parameter name
	Key    *string // Parameter key
	Value  *string // Parameter value
	Remark *string // Remark
}

// ExportInput defines input for Export function.
type ExportInput struct {
	Name      string // Parameter name, supports fuzzy search
	Key       string // Parameter key, supports fuzzy search
	BeginTime string // Creation time start
	EndTime   string // Creation time end
	Ids       []int  // Specific IDs to export; if empty, export all matching records
}
