// Package dynamicservice implements guest-side backend services for the
// plugin-demo-dynamic sample plugin.
package dynamicservice

import (
	"lina-core/pkg/plugindb"

	"lina-core/pkg/pluginbridge"
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
	// RegisterCrons publishes all built-in cron declarations for host-side
	// discovery.
	RegisterCrons() error
	// BuildCronHeartbeatPayload executes the declared cron heartbeat task and
	// returns a lightweight execution summary.
	BuildCronHeartbeatPayload() (*cronHeartbeatPayload, error)
}

// Interface compliance assertion for the default dynamic sample service
// implementation.
var _ Service = (*serviceImpl)(nil)

// runtimeHostService abstracts guest runtime host-call helpers used by the
// sample service.
type runtimeHostService interface {
	// Log writes one structured runtime log entry through the host.
	Log(level int, message string, fields map[string]string) error
	// StateGetInt reads one integer runtime state value.
	StateGetInt(key string) (int, bool, error)
	// StateSetInt writes one integer runtime state value.
	StateSetInt(key string, value int) error
	// Now returns the current host time string.
	Now() (string, error)
	// UUID returns one host-generated unique identifier string.
	UUID() (string, error)
	// Node returns the current host node identity string.
	Node() (string, error)
}

// storageHostService abstracts guest storage host-call helpers used by the
// sample service.
type storageHostService interface {
	// Put writes one governed storage object.
	Put(objectPath string, body []byte, contentType string, overwrite bool) (*pluginbridge.HostServiceStorageObject, error)
	// Get reads one governed storage object.
	Get(objectPath string) ([]byte, *pluginbridge.HostServiceStorageObject, bool, error)
	// Delete removes one governed storage object.
	Delete(objectPath string) error
	// List lists governed storage objects under one prefix.
	List(prefix string, limit uint32) ([]*pluginbridge.HostServiceStorageObject, error)
	// Stat reads metadata for one governed storage object.
	Stat(objectPath string) (*pluginbridge.HostServiceStorageObject, bool, error)
}

// networkHostService abstracts guest outbound HTTP host-call helpers used by
// the sample service.
type networkHostService interface {
	// Request executes one governed outbound HTTP request through the host.
	Request(targetURL string, request *pluginbridge.HostServiceNetworkRequest) (*pluginbridge.HostServiceNetworkResponse, error)
}

// cronHostService abstracts guest cron registration host-call helpers used by
// the sample service.
type cronHostService interface {
	// Register submits one built-in cron declaration for host-side discovery.
	Register(contract *pluginbridge.CronContract) error
}

// serviceImpl implements Service.
type serviceImpl struct {
	runtimeSvc runtimeHostService
	storageSvc storageHostService
	httpSvc    networkHostService
	cronSvc    cronHostService
	dataSvc    *plugindb.DB
}

// New creates and returns a new dynamic plugin backend service.
func New() Service {
	return &serviceImpl{
		runtimeSvc: newRuntimeHostService(),
		storageSvc: newStorageHostService(),
		httpSvc:    newNetworkHostService(),
		cronSvc:    newCronHostService(),
		dataSvc:    plugindb.Open(),
	}
}
