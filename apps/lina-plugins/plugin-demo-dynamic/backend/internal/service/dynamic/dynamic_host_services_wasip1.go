//go:build wasip1

// This file binds the sample plugin service to the guest-side pluginbridge
// host-call clients when compiling the plugin to Wasm.

package dynamicservice

import "lina-core/pkg/pluginbridge"

// newRuntimeHostService returns the real guest-side runtime host client.
func newRuntimeHostService() runtimeHostService {
	return pluginbridge.Runtime()
}

// newStorageHostService returns the real guest-side storage host client.
func newStorageHostService() storageHostService {
	return pluginbridge.Storage()
}

// newNetworkHostService returns the real guest-side outbound HTTP host client.
func newNetworkHostService() networkHostService {
	return pluginbridge.HTTP()
}

// newCronHostService returns the real guest-side cron registration host client.
func newCronHostService() cronHostService {
	return pluginbridge.Cron()
}
