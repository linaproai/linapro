// extidspi_manager.go owns the host-side provider manager and the
// lazy, enablement-gated Provider that the auth service binds. It mirrors the
// tenantspi/orgspi manager pattern: a source plugin registers a ProviderFactory
// at declaration time, the host holds the manager, and the consuming Provider
// returned by New resolves the single enabled provider on demand. When no
// provider plugin is enabled the returned Provider is fail-closed, which is the
// neutral value for external identity (reject, never mint accounts or sessions).
package extidspi

import (
	"context"

	"lina-core/pkg/plugin/capability/bizctxcap"
	internalregistry "lina-core/pkg/plugin/capability/internal/capabilityregistry"
	"lina-core/pkg/plugin/capability/usercap"
)

// CapabilityExternalIdentityV1 identifies the versioned external-identity
// framework capability in the shared provider registry.
const CapabilityExternalIdentityV1 = "external-identity.v1"

// ProviderEnv carries the explicit host services an external-identity provider
// adapter may use during lazy construction. It exposes only stable capability
// contracts, never host dao/do/entity, token minting, or session storage.
type ProviderEnv struct {
	// PluginID is the external-identity provider plugin being constructed.
	PluginID string
	// Users resolves host-owned user projections and the least-privilege
	// external-provisioning primitive without exposing sys_user storage.
	Users usercap.Service
	// BizCtx exposes the current request business context without host internals.
	BizCtx bizctxcap.Service
}

// ProviderFactory creates one external-identity provider from an explicit,
// typed construction environment during lazy capability use.
type ProviderFactory func(ctx context.Context, env ProviderEnv) (Provider, error)

// Manager owns external-identity provider declarations and lazy provider
// instances.
type Manager struct {
	registry *internalregistry.Manager[ProviderEnv]
}

// NewManager creates an empty external-identity provider manager.
func NewManager() *Manager {
	return &Manager{registry: internalregistry.NewManager[ProviderEnv]()}
}

// RegisterFactory records one plugin-provided external-identity capability
// factory.
func (m *Manager) RegisterFactory(pluginID string, factory ProviderFactory) error {
	return m.registry.RegisterFactory(
		CapabilityExternalIdentityV1,
		pluginID,
		func(ctx context.Context, env ProviderEnv) (any, error) {
			return factory(ctx, env)
		},
	)
}

// New creates the host-owned external-identity Provider. It resolves the single
// enabled plugin provider lazily on each call, gated by plugin enablement, so
// disabling the provider plugin immediately fails external login closed and
// re-enabling restores it. A nil manager, nil enablement, or no enabled
// provider yields a fail-closed Provider.
func New(
	manager *Manager,
	enablement internalregistry.EnablementReader,
	envFactory internalregistry.ProviderEnvFactory[ProviderEnv],
) Provider {
	if manager == nil {
		manager = NewManager()
	}
	if enablement == nil {
		enablement = disabledEnablementReader{}
	}
	if envFactory == nil {
		envFactory = defaultProviderEnv
	}
	return &managedProvider{
		manager:    manager,
		enablement: enablement,
		envFactory: envFactory,
	}
}

// managedProvider delegates external-identity calls to the active plugin
// provider and fails closed when none is usable.
type managedProvider struct {
	manager    *Manager
	enablement internalregistry.EnablementReader
	envFactory internalregistry.ProviderEnvFactory[ProviderEnv]
}

// Ensure managedProvider implements the external-identity Provider contract.
var _ Provider = (*managedProvider)(nil)

// active resolves the single enabled provider, or nil when none is usable.
func (p *managedProvider) active(ctx context.Context) (Provider, error) {
	if p == nil || p.manager == nil {
		return nil, nil
	}
	resolved, err := p.manager.registry.ActiveProviderWithError(
		ctx,
		CapabilityExternalIdentityV1,
		p.enablement,
		p.envFactory,
	)
	if err != nil {
		return nil, err
	}
	if resolved == nil {
		return nil, nil
	}
	provider, ok := resolved.(Provider)
	if !ok {
		return nil, nil
	}
	return provider, nil
}

// Resolve delegates to the active provider. No provider yields found=false so
// external login stays fail-closed without leaking account existence.
func (p *managedProvider) Resolve(ctx context.Context, in ResolveInput) (int, bool, error) {
	provider, err := p.active(ctx)
	if err != nil {
		return 0, false, err
	}
	if provider == nil {
		return 0, false, nil
	}
	return provider.Resolve(ctx, in)
}

// Provision delegates to the active provider. No provider fails closed with a
// not-provisioned error; the host never mints an account without a provider.
func (p *managedProvider) Provision(ctx context.Context, in ProvisionInput) (int, error) {
	provider, err := p.active(ctx)
	if err != nil {
		return 0, err
	}
	if provider == nil {
		return 0, ErrProviderUnavailable
	}
	return provider.Provision(ctx, in)
}

// Bind delegates to the active provider, failing closed when none is enabled.
func (p *managedProvider) Bind(ctx context.Context, in BindInput) error {
	provider, err := p.active(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return ErrProviderUnavailable
	}
	return provider.Bind(ctx, in)
}

// Unbind delegates to the active provider, failing closed when none is enabled.
func (p *managedProvider) Unbind(ctx context.Context, in UnbindInput) error {
	provider, err := p.active(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return ErrProviderUnavailable
	}
	return provider.Unbind(ctx, in)
}

// List delegates to the active provider, returning an empty set when none is
// enabled so callers see no bound identities rather than an error.
func (p *managedProvider) List(ctx context.Context, userID int) ([]BoundIdentity, error) {
	provider, err := p.active(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return []BoundIdentity{}, nil
	}
	return provider.List(ctx, userID)
}

// disabledEnablementReader reports every provider plugin as disabled.
type disabledEnablementReader struct{}

// IsProviderEnabled always returns false, keeping external login fail-closed
// until a real enablement reader is bound.
func (disabledEnablementReader) IsProviderEnabled(context.Context, string) bool {
	return false
}

// defaultProviderEnv creates a minimal provider environment when no host plugin
// runtime has been bound.
func defaultProviderEnv(_ context.Context, pluginID string) ProviderEnv {
	return ProviderEnv{PluginID: pluginID}
}
