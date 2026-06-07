// This file defines the shared guest host-service client contracts used by
// both wasip1 host-call implementations and non-WASI unsupported stubs.

package guest

import (
	"time"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// RuntimeHostService exposes guest-side helpers for the runtime host service.
type RuntimeHostService interface {
	// Log writes one structured runtime log entry through the host.
	Log(level int, message string, fields map[string]string) error
	// StateGet reads one plugin-scoped runtime state value by key.
	StateGet(key string) (string, bool, error)
	// StateSet writes one plugin-scoped runtime state value.
	StateSet(key string, value string) error
	// StateDelete removes one plugin-scoped runtime state value.
	StateDelete(key string) error
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

// StorageHostService exposes guest-side helpers for the governed storage host service.
type StorageHostService interface {
	// Put writes one governed storage object under the given logical path.
	Put(objectPath string, body []byte, contentType string, overwrite bool) (*protocol.HostServiceStorageObject, error)
	// PutText writes one UTF-8 text object under the given logical path.
	PutText(objectPath string, content string, contentType string, overwrite bool) (*protocol.HostServiceStorageObject, error)
	// Get reads one governed storage object under the given logical path.
	Get(objectPath string) ([]byte, *protocol.HostServiceStorageObject, bool, error)
	// GetText reads one UTF-8 text object under the given logical path.
	GetText(objectPath string) (string, *protocol.HostServiceStorageObject, bool, error)
	// Delete removes one governed storage object under the given logical path.
	Delete(objectPath string) error
	// List lists governed storage objects under one logical path prefix.
	List(prefix string, limit uint32) ([]*protocol.HostServiceStorageObject, error)
	// Stat reads metadata for one governed storage object under the given logical path.
	Stat(objectPath string) (*protocol.HostServiceStorageObject, bool, error)
}

// NetworkHostService exposes guest-side helpers for the governed outbound network host service.
type NetworkHostService interface {
	// Request executes one governed outbound HTTP request through the host.
	Request(targetURL string, request *protocol.HostServiceNetworkRequest) (*protocol.HostServiceNetworkResponse, error)
}

// CacheHostService exposes guest-side helpers for the governed distributed cache host service.
type CacheHostService interface {
	// Get reads one governed cache value from the authorized namespace.
	Get(namespace string, key string) (*protocol.HostServiceCacheValue, bool, error)
	// Set writes one governed cache value into the authorized namespace.
	Set(namespace string, key string, value string, expireSeconds int64) (*protocol.HostServiceCacheValue, error)
	// Delete removes one governed cache value from the authorized namespace.
	Delete(namespace string, key string) error
	// Incr increments one governed cache integer value inside the authorized namespace.
	Incr(namespace string, key string, delta int64, expireSeconds int64) (*protocol.HostServiceCacheValue, error)
	// Expire updates one governed cache expiration policy inside the authorized namespace.
	Expire(namespace string, key string, expireSeconds int64) (bool, string, error)
}

// LockHostService exposes guest-side helpers for the governed distributed lock host service.
type LockHostService interface {
	// Acquire attempts to acquire one governed distributed lock.
	Acquire(lockName string, leaseMillis int64) (*protocol.HostServiceLockAcquireResponse, error)
	// Renew extends one governed distributed lock using the issued ticket.
	Renew(lockName string, ticket string) (*protocol.HostServiceLockRenewResponse, error)
	// Release releases one governed distributed lock using the issued ticket.
	Release(lockName string, ticket string) error
}

// ConfigHostService exposes guest-side helpers for the read-only config host service.
type ConfigHostService interface {
	// Get reads one plugin-scoped configuration value as JSON.
	Get(key string) (string, bool, error)
	// Exists reports whether one configuration key exists.
	Exists(key string) (bool, error)
	// String reads one configuration value as a string.
	String(key string) (string, bool, error)
	// Bool reads one configuration value as a bool.
	Bool(key string) (bool, bool, error)
	// Int reads one configuration value as an int.
	Int(key string) (int, bool, error)
	// Duration reads one configuration value as a duration.
	Duration(key string) (time.Duration, bool, error)
}

// PluginService exposes guest-side plugin-domain host service clients.
type PluginService interface {
	// Config returns the plugin-scoped static config host service guest client.
	Config() ConfigHostService
}

// NotifyHostService exposes guest-side helpers for the governed unified notify host service.
type NotifyHostService interface {
	// Send sends one governed notification through the authorized channel.
	Send(channelKey string, request *protocol.HostServiceNotifySendRequest) (*protocol.HostServiceNotifySendResponse, error)
}

// CronHostService exposes guest-side helpers for the cron registration host
// service.
type CronHostService interface {
	// Register submits one dynamic-plugin cron declaration to the current
	// host-side discovery collector.
	Register(contract *protocol.CronContract) error
}

// HostConfigHostService exposes guest-side helpers for authorized host config reads.
type HostConfigHostService interface {
	// Get reads one authorized host config value as JSON.
	Get(key string) (string, bool, error)
	// String reads one authorized host config value as a string.
	String(key string) (string, bool, error)
	// Bool reads one authorized host config value as a bool.
	Bool(key string) (bool, bool, error)
	// Int reads one authorized host config value as an int.
	Int(key string) (int, bool, error)
	// Duration reads one authorized host config value as a duration.
	Duration(key string) (time.Duration, bool, error)
}

// ManifestHostService exposes guest-side helpers for plugin-scoped manifest resources.
type ManifestHostService interface {
	// Get reads one manifest resource as bytes.
	Get(path string) ([]byte, bool, error)
	// GetText reads one manifest resource as UTF-8 text.
	GetText(path string) (string, bool, error)
	// Scan decodes a YAML manifest resource or nested key into target.
	Scan(path string, key string, target any) (bool, error)
}
