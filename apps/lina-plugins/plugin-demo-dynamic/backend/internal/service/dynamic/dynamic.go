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
	BuildBackendSummaryPayload(request *pluginbridge.BridgeRequestEnvelopeV1) *backendSummaryPayload
	// BuildHostCallDemoPayload executes the host service demo and returns the
	// response payload.
	BuildHostCallDemoPayload(request *pluginbridge.BridgeRequestEnvelopeV1) (*hostCallDemoPayload, error)
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
