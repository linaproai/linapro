// Package capabilityregistry owns plugin capability provider declarations,
// lazy provider instances, fallback availability, and singleton conflict detection.
package capabilityregistry

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// ReasonNoProvider indicates no provider is currently active.
	ReasonNoProvider = "no_provider"
	// ReasonPluginDisabled indicates a declared provider plugin is not enabled.
	ReasonPluginDisabled = "plugin_disabled"
	// ReasonConflict indicates a singleton capability has more than one active provider.
	ReasonConflict = "provider_conflict"
	// ReasonProviderError indicates the enabled provider cannot be constructed.
	ReasonProviderError = "provider_error"
)

// ProviderFactory creates one provider instance for the supplied typed activation environment.
type ProviderFactory[Env any] func(ctx context.Context, env Env) (any, error)

// ProviderStatus describes one internal provider declaration state.
type ProviderStatus struct {
	// CapabilityID is the framework capability identifier.
	CapabilityID string
	// PluginID is the provider plugin identifier.
	PluginID string
	// Active reports whether this provider currently serves capability calls.
	Active bool
	// Conflict reports whether this provider is blocked by a singleton conflict.
	Conflict bool
	// Reason contains a stable diagnostic reason for inactive or conflicted state.
	Reason string
}

// CapabilityStatus describes the declared and currently usable provider state
// for one capability.
type CapabilityStatus struct {
	// CapabilityID is the framework capability identifier.
	CapabilityID string
	// Available reports whether the capability currently has an active provider.
	Available bool
	// ActiveProvider is the active provider plugin identifier, when available.
	ActiveProvider string
	// Reason contains a stable diagnostic reason when the capability is unavailable.
	Reason string
	// Providers contains all provider plugin states known for this capability.
	Providers []ProviderStatus
}

// EnablementReader reports whether one provider plugin is allowed to serve
// framework capability calls for the current request.
type EnablementReader interface {
	// IsProviderEnabled returns whether pluginID is currently platform-enabled as a provider.
	IsProviderEnabled(ctx context.Context, pluginID string) bool
}

// ProviderEnvFactory creates the lazy typed provider environment for a plugin.
type ProviderEnvFactory[Env any] func(ctx context.Context, pluginID string) Env

// Manager stores framework capability factories and lazy provider instances.
type Manager[Env any] struct {
	mu sync.RWMutex

	factories map[string]map[string]ProviderFactory[Env]
	providers map[string]map[string]any
}

// NewManager creates an empty capability manager.
func NewManager[Env any]() *Manager[Env] {
	return &Manager[Env]{
		factories: make(map[string]map[string]ProviderFactory[Env]),
		providers: make(map[string]map[string]any),
	}
}

// RegisterFactory records one plugin provider factory for a capability.
func (m *Manager[Env]) RegisterFactory(capabilityID string, pluginID string, factory ProviderFactory[Env]) error {
	if m == nil {
		return gerror.New("capability manager is nil")
	}
	capabilityID = strings.TrimSpace(capabilityID)
	pluginID = strings.TrimSpace(pluginID)
	if capabilityID == "" {
		return gerror.New("capability id is required")
	}
	if pluginID == "" {
		return gerror.New("capability provider plugin id is required")
	}
	if factory == nil {
		return gerror.Newf("capability provider factory is nil: capability=%s plugin=%s", capabilityID, pluginID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.factories[capabilityID] == nil {
		m.factories[capabilityID] = make(map[string]ProviderFactory[Env])
	}
	if _, exists := m.factories[capabilityID][pluginID]; exists {
		return gerror.Newf("capability provider factory already registered: capability=%s plugin=%s", capabilityID, pluginID)
	}
	m.factories[capabilityID][pluginID] = factory
	return nil
}

// ActiveProviderWithError returns the singleton enabled provider for a
// capability, lazily creating the provider instance from its registered factory.
func (m *Manager[Env]) ActiveProviderWithError(
	ctx context.Context,
	capabilityID string,
	enablement EnablementReader,
	envFactory ProviderEnvFactory[Env],
) (any, error) {
	if m == nil {
		return nil, gerror.New("capability manager is nil")
	}
	capabilityID = strings.TrimSpace(capabilityID)
	if capabilityID == "" {
		return nil, gerror.New("capability id is required")
	}
	enabledIDs := m.enabledProviderIDs(ctx, capabilityID, enablement)
	if len(enabledIDs) == 0 {
		return nil, nil
	}
	if len(enabledIDs) > 1 {
		return nil, bizerr.NewCode(
			capmodel.CodeCapabilityProviderConflict,
			bizerr.P("capability", capabilityID),
			bizerr.P("providerIds", strings.Join(enabledIDs, ",")),
		)
	}
	return m.providerForPlugin(ctx, capabilityID, enabledIDs[0], envFactory)
}

// Status returns the current state for one capability.
func (m *Manager[Env]) Status(
	ctx context.Context,
	capabilityID string,
	enablement EnablementReader,
) CapabilityStatus {
	if m == nil {
		return CapabilityStatus{CapabilityID: strings.TrimSpace(capabilityID), Reason: ReasonNoProvider}
	}
	return m.status(ctx, strings.TrimSpace(capabilityID), enablement)
}

// StatusWithProvider returns capability state after validating that the active
// provider can be lazily constructed from its typed environment.
func (m *Manager[Env]) StatusWithProvider(
	ctx context.Context,
	capabilityID string,
	enablement EnablementReader,
	envFactory ProviderEnvFactory[Env],
) CapabilityStatus {
	if m == nil {
		return CapabilityStatus{CapabilityID: strings.TrimSpace(capabilityID), Reason: ReasonNoProvider}
	}
	status := m.status(ctx, strings.TrimSpace(capabilityID), enablement)
	if !status.Available || status.ActiveProvider == "" {
		return status
	}
	provider, err := m.providerForPlugin(ctx, status.CapabilityID, status.ActiveProvider, envFactory)
	if err == nil && provider != nil {
		return status
	}
	status.Available = false
	status.ActiveProvider = ""
	status.Reason = ReasonProviderError
	for i := range status.Providers {
		if status.Providers[i].Active {
			status.Providers[i].Active = false
			status.Providers[i].Reason = ReasonProviderError
		}
	}
	return status
}

// Statuses returns every known provider state sorted for deterministic scans.
func (m *Manager[Env]) Statuses(ctx context.Context, enablement EnablementReader) []ProviderStatus {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	capabilityIDs := make([]string, 0, len(m.factories))
	for capabilityID := range m.factories {
		capabilityIDs = append(capabilityIDs, capabilityID)
	}
	m.mu.RUnlock()
	sort.Strings(capabilityIDs)

	statuses := make([]ProviderStatus, 0)
	for _, capabilityID := range capabilityIDs {
		statuses = append(statuses, m.status(ctx, capabilityID, enablement).Providers...)
	}
	return statuses
}

// enabledProviderIDs returns deterministic provider plugin IDs currently
// allowed by plugin state.
func (m *Manager[Env]) enabledProviderIDs(
	ctx context.Context,
	capabilityID string,
	enablement EnablementReader,
) []string {
	if m == nil || enablement == nil {
		return []string{}
	}
	capabilityID = strings.TrimSpace(capabilityID)
	m.mu.RLock()
	providerIDs := sortedProviderIDs(m.factories[capabilityID])
	m.mu.RUnlock()

	enabledIDs := make([]string, 0, len(providerIDs))
	for _, pluginID := range providerIDs {
		if enablement.IsProviderEnabled(ctx, pluginID) {
			enabledIDs = append(enabledIDs, pluginID)
		}
	}
	return enabledIDs
}

// providerForPlugin returns or creates one provider instance.
func (m *Manager[Env]) providerForPlugin(
	ctx context.Context,
	capabilityID string,
	pluginID string,
	envFactory ProviderEnvFactory[Env],
) (any, error) {
	m.mu.RLock()
	if provider := m.providers[capabilityID][pluginID]; provider != nil {
		m.mu.RUnlock()
		return provider, nil
	}
	factory := m.factories[capabilityID][pluginID]
	m.mu.RUnlock()
	if factory == nil {
		return nil, nil
	}
	var env Env
	if envFactory != nil {
		env = envFactory(ctx, pluginID)
	}
	provider, err := factory(ctx, env)
	if err != nil {
		return nil, gerror.Wrapf(err, "create capability provider failed: capability=%s plugin=%s", capabilityID, pluginID)
	}
	if provider == nil {
		return nil, gerror.Newf("capability provider factory returned nil: capability=%s plugin=%s", capabilityID, pluginID)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.providers[capabilityID] == nil {
		m.providers[capabilityID] = make(map[string]any)
	}
	if cached := m.providers[capabilityID][pluginID]; cached != nil {
		return cached, nil
	}
	m.providers[capabilityID][pluginID] = provider
	return provider, nil
}

// status builds one capability status from declarations and plugin enablement.
func (m *Manager[Env]) status(
	ctx context.Context,
	capabilityID string,
	enablement EnablementReader,
) CapabilityStatus {
	status := CapabilityStatus{
		CapabilityID: capabilityID,
		Reason:       ReasonNoProvider,
		Providers:    make([]ProviderStatus, 0),
	}
	m.mu.RLock()
	providerIDs := sortedProviderIDs(m.factories[capabilityID])
	m.mu.RUnlock()

	enabledIDs := make(map[string]bool)
	for _, pluginID := range providerIDs {
		if enablement != nil && enablement.IsProviderEnabled(ctx, pluginID) {
			enabledIDs[pluginID] = true
		}
	}
	conflict := len(enabledIDs) > 1
	for _, pluginID := range providerIDs {
		enabled := enabledIDs[pluginID]
		providerStatus := ProviderStatus{
			CapabilityID: capabilityID,
			PluginID:     pluginID,
			Active:       enabled && !conflict,
			Conflict:     enabled && conflict,
		}
		if !enabled {
			providerStatus.Reason = ReasonPluginDisabled
		}
		if providerStatus.Active {
			status.Available = true
			status.ActiveProvider = pluginID
			status.Reason = ""
		}
		if providerStatus.Conflict {
			providerStatus.Reason = ReasonConflict
			status.Reason = ReasonConflict
		}
		status.Providers = append(status.Providers, providerStatus)
	}
	return status
}

// sortedProviderIDs returns deterministic provider IDs from one provider map.
func sortedProviderIDs[T any](providers map[string]T) []string {
	ids := make([]string, 0, len(providers))
	for pluginID := range providers {
		ids = append(ids, pluginID)
	}
	sort.Strings(ids)
	return ids
}
