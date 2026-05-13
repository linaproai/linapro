// Package demo implements public demo services exposed by the plugin-demo-source
// backend.
package demo

import (
	"context"

	plugincontract "lina-core/pkg/pluginservice/contract"
)

// Service defines the demo service contract.
type Service interface {
	// Ping returns the public ping payload used by route verification.
	Ping(ctx context.Context) (out *PingOutput, err error)
	// Summary returns the concise backend summary rendered on the plugin page.
	Summary(ctx context.Context) (out *SummaryOutput, err error)
	// ListRecords returns the paged demo records rendered by the source-plugin CRUD page.
	ListRecords(ctx context.Context, in *ListRecordsInput) (out *ListRecordsOutput, err error)
	// GetRecord returns one demo record detail for edit forms.
	GetRecord(ctx context.Context, id int64) (out *RecordDetailOutput, err error)
	// CreateRecord creates one demo record and stores its optional attachment file.
	CreateRecord(ctx context.Context, in *CreateRecordInput) (out *RecordMutationOutput, err error)
	// UpdateRecord updates one demo record and replaces or removes its optional attachment.
	UpdateRecord(ctx context.Context, in *UpdateRecordInput) (out *RecordMutationOutput, err error)
	// DeleteRecord deletes one demo record and cleans its attachment file.
	DeleteRecord(ctx context.Context, id int64) error
	// BuildAttachmentDownload returns one attachment download descriptor for the given record.
	BuildAttachmentDownload(ctx context.Context, id int64) (out *AttachmentDownloadOutput, err error)
	// PurgeStorageData clears plugin-owned attachment files before uninstall SQL drops the data table.
	PurgeStorageData(ctx context.Context) error
}

// Interface compliance assertion for the default demo service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	i18nSvc      plugincontract.I18nService         // i18nSvc resolves plugin runtime translations.
	tenantFilter plugincontract.TenantFilterService // tenantFilter constrains plugin-owned demo rows.
}

// New creates and returns a new demo service instance.
func New(i18nSvc plugincontract.I18nService, tenantFilter plugincontract.TenantFilterService) Service {
	return &serviceImpl{
		i18nSvc:      i18nSvc,
		tenantFilter: tenantFilter,
	}
}
