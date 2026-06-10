// This file implements process-local storage provider registration and active
// provider selection. Registration is intentionally explicit because storage is
// durable state and must not silently move across providers.

package storagecap

import (
	"context"
	"sort"
	"strings"
	"sync"

	"lina-core/pkg/bizerr"
)

var providerRegistry = struct {
	sync.RWMutex
	factories map[string]ProviderFactory
}{
	factories: make(map[string]ProviderFactory),
}

// Provide registers one plugin-provided storage provider factory. Provider IDs
// must be stable plugin IDs; LocalProviderID is reserved for the built-in local
// provider configured by the host service adapter.
func Provide(pluginID string, factory ProviderFactory) error {
	normalizedID := strings.TrimSpace(pluginID)
	if normalizedID == "" {
		return bizerr.NewCode(CodeStorageProviderIDRequired)
	}
	if factory == nil {
		return bizerr.NewCode(CodeStorageProviderFactoryRequired)
	}

	providerRegistry.Lock()
	defer providerRegistry.Unlock()
	if _, exists := providerRegistry.factories[normalizedID]; exists {
		return bizerr.NewCode(CodeStorageProviderAlreadyRegistered, bizerr.P("providerId", normalizedID))
	}
	providerRegistry.factories[normalizedID] = factory
	return nil
}

// ProviderFactoryFor returns one registered provider factory by provider ID.
func ProviderFactoryFor(providerID string) (ProviderFactory, bool) {
	providerRegistry.RLock()
	defer providerRegistry.RUnlock()
	factory, ok := providerRegistry.factories[strings.TrimSpace(providerID)]
	return factory, ok
}

// RegisteredProviderIDs returns all plugin-registered provider IDs.
func RegisteredProviderIDs() []string {
	providerRegistry.RLock()
	defer providerRegistry.RUnlock()
	ids := make([]string, 0, len(providerRegistry.factories))
	for id := range providerRegistry.factories {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// ResolveProvider selects and constructs the active provider. Empty active
// provider ID always selects localProvider. Non-empty active provider IDs must
// refer to an available registered plugin provider; failures do not fall back to
// local storage.
func ResolveProvider(
	ctx context.Context,
	runtime ProviderRuntime,
	localProvider Provider,
) (string, Provider, error) {
	activeID := ""
	if runtime != nil {
		activeID = strings.TrimSpace(runtime.ActiveProviderPluginID(ctx))
	}
	if activeID == "" {
		if localProvider == nil {
			return "", nil, bizerr.NewCode(CodeStorageProviderUnavailable)
		}
		return LocalProviderID, localProvider, nil
	}
	if runtime != nil && !runtime.ProviderPluginAvailable(ctx, activeID) {
		return activeID, nil, bizerr.NewCode(CodeStorageProviderUnavailable, bizerr.P("providerId", activeID))
	}
	factory, ok := ProviderFactoryFor(activeID)
	if !ok {
		return activeID, nil, bizerr.NewCode(CodeStorageProviderUnavailable, bizerr.P("providerId", activeID))
	}
	provider, err := factory(ctx, ProviderEnv{ProviderID: activeID, Runtime: runtime})
	if err != nil {
		return activeID, nil, err
	}
	if provider == nil {
		return activeID, nil, bizerr.NewCode(CodeStorageProviderUnavailable, bizerr.P("providerId", activeID))
	}
	return activeID, provider, nil
}

// ProviderStatuses returns active and availability snapshots for the built-in
// local provider plus plugin-registered providers.
func ProviderStatuses(ctx context.Context, runtime ProviderRuntime, localProvider Provider) []*ProviderStatus {
	activeID := ""
	if runtime != nil {
		activeID = strings.TrimSpace(runtime.ActiveProviderPluginID(ctx))
	}
	if activeID == "" {
		activeID = LocalProviderID
	}
	statuses := []*ProviderStatus{{
		ProviderID: LocalProviderID,
		Active:     activeID == LocalProviderID,
		Available:  localProvider != nil,
	}}
	if localProvider == nil {
		statuses[0].Message = "local provider is not configured"
	}
	for _, id := range RegisteredProviderIDs() {
		available := runtime == nil || runtime.ProviderPluginAvailable(ctx, id)
		status := &ProviderStatus{
			ProviderID: id,
			Active:     activeID == id,
			Available:  available,
		}
		if !available {
			status.Message = "provider plugin is not available"
		}
		statuses = append(statuses, status)
	}
	return statuses
}
