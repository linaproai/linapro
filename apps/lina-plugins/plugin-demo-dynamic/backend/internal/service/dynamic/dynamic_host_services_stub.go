//go:build !wasip1

// This file provides host-build stubs for the sample plugin host-call clients so
// the package still compiles during repository-wide go test runs.

package dynamicservice

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
)

// errDynamicHostCallsUnavailable reports that host-call clients are unavailable
// outside guest Wasm builds.
var errDynamicHostCallsUnavailable = gerror.New("plugin-demo-dynamic host calls are only available for wasip1 builds")

// unsupportedRuntimeHostService is the host-build stub runtime client.
type unsupportedRuntimeHostService struct{}

// newRuntimeHostService returns the host-build stub runtime client.
func newRuntimeHostService() runtimeHostService {
	return unsupportedRuntimeHostService{}
}

// Log reports that guest-only runtime host calls are unavailable in host builds.
func (unsupportedRuntimeHostService) Log(_ int, _ string, _ map[string]string) error {
	return errDynamicHostCallsUnavailable
}

// StateGetInt reports that guest-only runtime host calls are unavailable in host builds.
func (unsupportedRuntimeHostService) StateGetInt(_ string) (int, bool, error) {
	return 0, false, errDynamicHostCallsUnavailable
}

// StateSetInt reports that guest-only runtime host calls are unavailable in host builds.
func (unsupportedRuntimeHostService) StateSetInt(_ string, _ int) error {
	return errDynamicHostCallsUnavailable
}

// Now reports that guest-only runtime host calls are unavailable in host builds.
func (unsupportedRuntimeHostService) Now() (string, error) {
	return "", errDynamicHostCallsUnavailable
}

// UUID reports that guest-only runtime host calls are unavailable in host builds.
func (unsupportedRuntimeHostService) UUID() (string, error) {
	return "", errDynamicHostCallsUnavailable
}

// Node reports that guest-only runtime host calls are unavailable in host builds.
func (unsupportedRuntimeHostService) Node() (string, error) {
	return "", errDynamicHostCallsUnavailable
}

// unsupportedStorageHostService is the host-build stub storage client.
type unsupportedStorageHostService struct{}

// newStorageHostService returns the host-build stub storage client.
func newStorageHostService() storageHostService {
	return unsupportedStorageHostService{}
}

// Put reports that guest-only storage host calls are unavailable in host builds.
func (unsupportedStorageHostService) Put(
	_ string,
	_ []byte,
	_ string,
	_ bool,
) (*pluginbridge.HostServiceStorageObject, error) {
	return nil, errDynamicHostCallsUnavailable
}

// Get reports that guest-only storage host calls are unavailable in host builds.
func (unsupportedStorageHostService) Get(
	_ string,
) ([]byte, *pluginbridge.HostServiceStorageObject, bool, error) {
	return nil, nil, false, errDynamicHostCallsUnavailable
}

// Delete reports that guest-only storage host calls are unavailable in host builds.
func (unsupportedStorageHostService) Delete(_ string) error {
	return errDynamicHostCallsUnavailable
}

// List reports that guest-only storage host calls are unavailable in host builds.
func (unsupportedStorageHostService) List(
	_ string,
	_ uint32,
) ([]*pluginbridge.HostServiceStorageObject, error) {
	return nil, errDynamicHostCallsUnavailable
}

// Stat reports that guest-only storage host calls are unavailable in host builds.
func (unsupportedStorageHostService) Stat(
	_ string,
) (*pluginbridge.HostServiceStorageObject, bool, error) {
	return nil, false, errDynamicHostCallsUnavailable
}

// unsupportedNetworkHostService is the host-build stub network client.
type unsupportedNetworkHostService struct{}

// newNetworkHostService returns the host-build stub network client.
func newNetworkHostService() networkHostService {
	return unsupportedNetworkHostService{}
}

// Request reports that guest-only network host calls are unavailable in host builds.
func (unsupportedNetworkHostService) Request(
	_ string,
	_ *pluginbridge.HostServiceNetworkRequest,
) (*pluginbridge.HostServiceNetworkResponse, error) {
	return nil, errDynamicHostCallsUnavailable
}

// unsupportedCronHostService is the host-build stub cron registration client.
type unsupportedCronHostService struct{}

// newCronHostService returns the host-build stub cron registration client.
func newCronHostService() cronHostService {
	return unsupportedCronHostService{}
}

// Register reports that guest-only cron host calls are unavailable in host builds.
func (unsupportedCronHostService) Register(_ *pluginbridge.CronContract) error {
	return errDynamicHostCallsUnavailable
}
