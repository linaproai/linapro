//go:build !wasip1

// This file provides host-build stubs for guest-side host-service transport and
// clients. The stubs keep ordinary Go builds and unit tests compilable while
// making it explicit that real host calls are only available in wasip1 guest
// builds.

package pluginbridge

import (
	"context"
	"time"

	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// Host-build unsupported client implementations used by package-level helpers.
type (
	unsupportedRuntimeHostService    struct{}
	unsupportedStorageHostService    struct{}
	unsupportedNetworkHostService    struct{}
	unsupportedCacheHostService      struct{}
	unsupportedLockHostService       struct{}
	unsupportedHostConfigHostService struct{}
	unsupportedManifestHostService   struct{}
)

var (
	defaultRuntimeHostService    RuntimeHostService    = unsupportedRuntimeHostService{}
	defaultStorageHostService    storagecap.Service    = unsupportedStorageHostService{}
	defaultNetworkHostService    NetworkHostService    = unsupportedNetworkHostService{}
	defaultCacheHostService      cachecap.Service      = unsupportedCacheHostService{}
	defaultLockHostService       lockcap.Service       = unsupportedLockHostService{}
	defaultHostConfigHostService HostConfigHostService = unsupportedHostConfigHostService{}
	defaultManifestHostService   ManifestHostService   = unsupportedManifestHostService{}
)

// InvokeHostService reports that generic guest host calls are unavailable.
func InvokeHostService(_ string, _ string, _ string, _ string, _ []byte) ([]byte, error) {
	return nil, ErrHostCallsUnavailable
}

// Runtime returns the runtime host service guest client.
func Runtime() RuntimeHostService {
	return defaultRuntimeHostService
}

// Storage returns the storage domain guest client.
func Storage() storagecap.Service {
	return defaultStorageHostService
}

// Network returns the outbound network host service guest client.
func Network() NetworkHostService {
	return defaultNetworkHostService
}

// Cache returns the distributed cache domain guest client.
func Cache() cachecap.Service {
	return defaultCacheHostService
}

// Lock returns the distributed lock domain guest client.
func Lock() lockcap.Service {
	return defaultLockHostService
}

// HostConfig returns the host config guest client.
func HostConfig() HostConfigHostService {
	return defaultHostConfigHostService
}

// Manifest returns the plugin manifest-resource guest client.
func Manifest() ManifestHostService {
	return defaultManifestHostService
}

// Log reports that guest runtime host calls are unavailable.
func (unsupportedRuntimeHostService) Log(_ int, _ string, _ map[string]string) error {
	return ErrHostCallsUnavailable
}

// StateGet reports that guest runtime host calls are unavailable.
func (unsupportedRuntimeHostService) StateGet(_ string) (string, bool, error) {
	return "", false, ErrHostCallsUnavailable
}

// StateSet reports that guest runtime host calls are unavailable.
func (unsupportedRuntimeHostService) StateSet(_ string, _ string) error {
	return ErrHostCallsUnavailable
}

// StateDelete reports that guest runtime host calls are unavailable.
func (unsupportedRuntimeHostService) StateDelete(_ string) error {
	return ErrHostCallsUnavailable
}

// StateGetInt reports that guest runtime host calls are unavailable.
func (unsupportedRuntimeHostService) StateGetInt(_ string) (int, bool, error) {
	return 0, false, ErrHostCallsUnavailable
}

// StateSetInt reports that guest runtime host calls are unavailable.
func (unsupportedRuntimeHostService) StateSetInt(_ string, _ int) error {
	return ErrHostCallsUnavailable
}

// Now reports that guest runtime host calls are unavailable.
func (unsupportedRuntimeHostService) Now() (string, error) {
	return "", ErrHostCallsUnavailable
}

// UUID reports that guest runtime host calls are unavailable.
func (unsupportedRuntimeHostService) UUID() (string, error) {
	return "", ErrHostCallsUnavailable
}

// Node reports that guest runtime host calls are unavailable.
func (unsupportedRuntimeHostService) Node() (string, error) {
	return "", ErrHostCallsUnavailable
}

// HostLog writes one runtime log entry through the host.
func HostLog(level int, message string, fields map[string]string) error {
	return Runtime().Log(level, message, fields)
}

// HostStateGet reads one plugin-scoped runtime state value.
func HostStateGet(key string) (string, bool, error) {
	return Runtime().StateGet(key)
}

// HostStateSet writes one plugin-scoped runtime state value.
func HostStateSet(key string, value string) error {
	return Runtime().StateSet(key, value)
}

// HostStateDelete removes one plugin-scoped runtime state value.
func HostStateDelete(key string) error {
	return Runtime().StateDelete(key)
}

// HostStateGetInt reads one integer plugin-scoped runtime state value.
func HostStateGetInt(key string) (int, bool, error) {
	return Runtime().StateGetInt(key)
}

// HostStateSetInt writes one integer plugin-scoped runtime state value.
func HostStateSetInt(key string, value int) error {
	return Runtime().StateSetInt(key, value)
}

// Put reports that guest storage host calls are unavailable.
func (unsupportedStorageHostService) Put(_ context.Context, _ storagecap.PutInput) (*storagecap.PutOutput, error) {
	return nil, ErrHostCallsUnavailable
}

// Get reports that guest storage host calls are unavailable.
func (unsupportedStorageHostService) Get(_ context.Context, _ storagecap.GetInput) (*storagecap.GetOutput, error) {
	return nil, ErrHostCallsUnavailable
}

// Delete reports that guest storage host calls are unavailable.
func (unsupportedStorageHostService) Delete(_ context.Context, _ storagecap.DeleteInput) error {
	return ErrHostCallsUnavailable
}

// List reports that guest storage host calls are unavailable.
func (unsupportedStorageHostService) List(_ context.Context, _ storagecap.ListInput) (*storagecap.ListOutput, error) {
	return nil, ErrHostCallsUnavailable
}

// Stat reports that guest storage host calls are unavailable.
func (unsupportedStorageHostService) Stat(_ context.Context, _ storagecap.StatInput) (*storagecap.StatOutput, error) {
	return nil, ErrHostCallsUnavailable
}

// ProviderStatuses reports that guest storage host calls are unavailable.
func (unsupportedStorageHostService) ProviderStatuses(_ context.Context) ([]*storagecap.ProviderStatus, error) {
	return nil, ErrHostCallsUnavailable
}

// Request reports that guest outbound network host calls are unavailable.
func (unsupportedNetworkHostService) Request(
	_ string,
	_ *protocol.HostServiceNetworkRequest,
) (*protocol.HostServiceNetworkResponse, error) {
	return nil, ErrHostCallsUnavailable
}

// Get reports that guest cache host calls are unavailable.
func (unsupportedCacheHostService) Get(_ context.Context, _ string, _ string) (*cachecap.CacheItem, bool, error) {
	return nil, false, ErrHostCallsUnavailable
}

// Set reports that guest cache host calls are unavailable.
func (unsupportedCacheHostService) Set(_ context.Context, _ string, _ string, _ string, _ time.Duration) (*cachecap.CacheItem, error) {
	return nil, ErrHostCallsUnavailable
}

// Delete reports that guest cache host calls are unavailable.
func (unsupportedCacheHostService) Delete(_ context.Context, _ string, _ string) error {
	return ErrHostCallsUnavailable
}

// Incr reports that guest cache host calls are unavailable.
func (unsupportedCacheHostService) Incr(_ context.Context, _ string, _ string, _ int64, _ time.Duration) (*cachecap.CacheItem, error) {
	return nil, ErrHostCallsUnavailable
}

// Expire reports that guest cache host calls are unavailable.
func (unsupportedCacheHostService) Expire(_ context.Context, _ string, _ string, _ time.Duration) (bool, *time.Time, error) {
	return false, nil, ErrHostCallsUnavailable
}

// Acquire reports that guest lock host calls are unavailable.
func (unsupportedLockHostService) Acquire(_ context.Context, _ lockcap.AcquireInput) (*lockcap.AcquireOutput, error) {
	return nil, ErrHostCallsUnavailable
}

// Renew reports that guest lock host calls are unavailable.
func (unsupportedLockHostService) Renew(_ context.Context, _ lockcap.RenewInput) (*lockcap.RenewOutput, error) {
	return nil, ErrHostCallsUnavailable
}

// Release reports that guest lock host calls are unavailable.
func (unsupportedLockHostService) Release(_ context.Context, _ lockcap.ReleaseInput) error {
	return ErrHostCallsUnavailable
}

// configValue reports that guest plugin config host calls are unavailable.
func configValue(_ string) (string, bool, error) {
	return "", false, ErrHostCallsUnavailable
}

// Get reports that guest host config calls are unavailable.
func (unsupportedHostConfigHostService) Get(_ string) (string, bool, error) {
	return "", false, ErrHostCallsUnavailable
}

// String reports that guest host config calls are unavailable.
func (unsupportedHostConfigHostService) String(_ string) (string, bool, error) {
	return "", false, ErrHostCallsUnavailable
}

// Bool reports that guest host config calls are unavailable.
func (unsupportedHostConfigHostService) Bool(_ string) (bool, bool, error) {
	return false, false, ErrHostCallsUnavailable
}

// Int reports that guest host config calls are unavailable.
func (unsupportedHostConfigHostService) Int(_ string) (int, bool, error) {
	return 0, false, ErrHostCallsUnavailable
}

// Duration reports that guest host config calls are unavailable.
func (unsupportedHostConfigHostService) Duration(_ string) (time.Duration, bool, error) {
	return 0, false, ErrHostCallsUnavailable
}

// Get reports that guest manifest host calls are unavailable.
func (unsupportedManifestHostService) Get(_ string) ([]byte, bool, error) {
	return nil, false, ErrHostCallsUnavailable
}

// GetText reports that guest manifest host calls are unavailable.
func (unsupportedManifestHostService) GetText(_ string) (string, bool, error) {
	return "", false, ErrHostCallsUnavailable
}

// Scan reports that guest manifest host calls are unavailable.
func (unsupportedManifestHostService) Scan(_ string, _ string, _ any) (bool, error) {
	return false, ErrHostCallsUnavailable
}
