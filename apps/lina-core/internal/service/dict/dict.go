// Package dict implements dictionary type, dictionary data, and import/export
// services for the Lina core host service.
package dict

import (
	"context"
	"io"

	"lina-core/internal/model/entity"
)

// Service defines the complete dict service contract.
type Service interface {
	TypeService
	DataService
	ImportExportService
	LookupService
}

// TypeService defines dictionary type management operations.
type TypeService interface {
	// List queries dictionary types with pagination and filters.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// Create creates a new dictionary type.
	Create(ctx context.Context, in CreateInput) (int, error)
	// GetById retrieves a dictionary type by ID.
	GetById(ctx context.Context, id int) (*entity.SysDictType, error)
	// Update updates an existing dictionary type.
	Update(ctx context.Context, in UpdateInput) error
	// Delete deletes a dictionary type by ID.
	Delete(ctx context.Context, id int) error
	// Options returns enabled dictionary type options for selectors.
	Options(ctx context.Context) ([]*OptionItem, error)
}

// DataService defines dictionary data management operations.
type DataService interface {
	// DataList queries dictionary data entries with pagination and filters.
	DataList(ctx context.Context, in DataListInput) (*DataListOutput, error)
	// DataCreate creates a new dictionary data entry.
	DataCreate(ctx context.Context, in DataCreateInput) (int, error)
	// DataGetById retrieves a dictionary data entry by ID.
	DataGetById(ctx context.Context, id int) (*entity.SysDictData, error)
	// DataUpdate updates an existing dictionary data entry.
	DataUpdate(ctx context.Context, in DataUpdateInput) error
	// DataDelete deletes a dictionary data entry by ID.
	DataDelete(ctx context.Context, id int) error
	// DataByType lists enabled dictionary data entries for one dictionary type.
	DataByType(ctx context.Context, dictType string) ([]*entity.SysDictData, error)
}

// ImportExportService defines dictionary workbook import and export operations.
type ImportExportService interface {
	// Export exports dictionary types to an Excel file.
	Export(ctx context.Context, in ExportInput) (data []byte, err error)
	// DataExport exports dictionary data entries to an Excel file.
	DataExport(ctx context.Context, in DataExportInput) (data []byte, err error)
	// CombinedExport exports dictionary types and data together to an Excel file.
	CombinedExport(ctx context.Context, in CombinedExportInput) (data []byte, err error)
	// CombinedImport imports dictionary types and data from a combined workbook.
	CombinedImport(ctx context.Context, fileData []byte, updateSupport bool) (result *CombinedImportResult, err error)
	// CombinedImportTemplate generates the combined dictionary import template.
	CombinedImportTemplate(ctx context.Context) (data []byte, err error)
	// TypeImport imports dictionary types from an Excel reader.
	TypeImport(ctx context.Context, file io.Reader, updateSupport bool) (result *ImportResult, err error)
	// DataImport imports dictionary data entries from an Excel reader.
	DataImport(ctx context.Context, file io.Reader, updateSupport bool) (result *ImportResult, err error)
	// GenerateTypeImportTemplate generates the dictionary type import template.
	GenerateTypeImportTemplate(ctx context.Context) (data []byte, err error)
	// GenerateDataImportTemplate generates the dictionary data import template.
	GenerateDataImportTemplate(ctx context.Context) (data []byte, err error)
}

// LookupService defines dictionary label lookup and map building operations.
type LookupService interface {
	// GetLabelByValue returns a dictionary label by dictionary type and string value.
	GetLabelByValue(ctx context.Context, in GetLabelByValueInput) string
	// GetLabelByIntValue returns a dictionary label by dictionary type and integer value.
	GetLabelByIntValue(ctx context.Context, dictType string, value int) string
	// BuildLabelMap builds a string-value-to-label map for one dictionary type.
	BuildLabelMap(ctx context.Context, dictType string) map[string]string
	// BuildIntLabelMap builds an integer-value-to-label map for one dictionary type.
	BuildIntLabelMap(ctx context.Context, dictType string) map[int]string
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	i18nSvc dictI18nTranslator
}

// New creates a dict service from explicit runtime-owned dependencies.
func New(i18nSvc dictI18nTranslator) Service {
	return &serviceImpl{
		i18nSvc: i18nSvc,
	}
}
