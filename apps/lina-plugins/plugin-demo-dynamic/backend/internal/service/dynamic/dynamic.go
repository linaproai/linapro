// Package dynamicservice implements guest-side backend services for the
// plugin-demo-dynamic sample plugin.
package dynamicservice

import (
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/plugindb"
)

// Service defines the dynamic service contract.
type Service interface {
	// BuildBackendSummaryPayload builds the backend summary response payload.
	BuildBackendSummaryPayload(input *BackendSummaryInput) *backendSummaryPayload
	// ListDemoRecordsPayload returns one paged demo-record list backed by the
	// plugin-owned SQL table.
	ListDemoRecordsPayload(input *DemoRecordListInput) (*demoRecordListPayload, error)
	// GetDemoRecordPayload returns one demo-record detail by ID.
	GetDemoRecordPayload(recordID string) (*demoRecordPayload, error)
	// CreateDemoRecordPayload creates one demo record and stores its optional attachment.
	CreateDemoRecordPayload(input *DemoRecordMutationInput) (*demoRecordPayload, error)
	// UpdateDemoRecordPayload updates one demo record and replaces or removes its optional attachment.
	UpdateDemoRecordPayload(recordID string, input *DemoRecordMutationInput) (*demoRecordPayload, error)
	// DeleteDemoRecordPayload deletes one demo record and its optional attachment.
	DeleteDemoRecordPayload(recordID string) (*demoRecordDeletePayload, error)
	// BuildDemoRecordAttachmentDownload returns one attachment download descriptor.
	BuildDemoRecordAttachmentDownload(recordID string) (*demoRecordAttachmentDownloadPayload, error)
	// BuildHostCallDemoPayload executes the host service demo and returns the
	// response payload.
	BuildHostCallDemoPayload(input *HostCallDemoInput) (*hostCallDemoPayload, error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	runtimeSvc pluginbridge.RuntimeHostService
	storageSvc pluginbridge.StorageHostService
	httpSvc    pluginbridge.HTTPHostService
	dataSvc    *plugindb.DB
}

// New creates and returns a new dynamic plugin backend service.
func New() Service {
	return &serviceImpl{
		runtimeSvc: pluginbridge.Runtime(),
		storageSvc: pluginbridge.Storage(),
		httpSvc:    pluginbridge.HTTP(),
		dataSvc:    plugindb.Open(),
	}
}
