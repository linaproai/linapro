// Package media implements strategy, binding, and stream-alias services exposed by the media plugin.
package media

import (
	"context"

	"lina-core/pkg/pluginservice/bizctx"
)

// Service defines the media plugin service contract.
type Service interface {
	// ListStrategies returns paged media strategies.
	ListStrategies(ctx context.Context, in ListStrategiesInput) (*ListStrategiesOutput, error)
	// GetStrategy returns one media strategy by ID.
	GetStrategy(ctx context.Context, id int64) (*StrategyOutput, error)
	// CreateStrategy creates one media strategy.
	CreateStrategy(ctx context.Context, in StrategyMutationInput) (int64, error)
	// UpdateStrategy updates one media strategy.
	UpdateStrategy(ctx context.Context, id int64, in StrategyMutationInput) error
	// UpdateStrategyEnable changes one media strategy enable status.
	UpdateStrategyEnable(ctx context.Context, id int64, enable int) error
	// SetGlobalStrategy sets one media strategy as the active global strategy.
	SetGlobalStrategy(ctx context.Context, id int64) error
	// DeleteStrategy deletes one unreferenced media strategy.
	DeleteStrategy(ctx context.Context, id int64) error
	// ListDeviceBindings returns paged device strategy bindings.
	ListDeviceBindings(ctx context.Context, in ListBindingsInput) (*ListBindingsOutput, error)
	// SaveDeviceBinding creates or updates one device strategy binding.
	SaveDeviceBinding(ctx context.Context, in DeviceBindingMutationInput) (*DeviceBindingMutationOutput, error)
	// DeleteDeviceBinding deletes one device strategy binding.
	DeleteDeviceBinding(ctx context.Context, deviceID string) (*DeviceBindingMutationOutput, error)
	// ListTenantBindings returns paged tenant strategy bindings.
	ListTenantBindings(ctx context.Context, in ListBindingsInput) (*ListBindingsOutput, error)
	// SaveTenantBinding creates or updates one tenant strategy binding.
	SaveTenantBinding(ctx context.Context, in TenantBindingMutationInput) (*TenantBindingMutationOutput, error)
	// DeleteTenantBinding deletes one tenant strategy binding.
	DeleteTenantBinding(ctx context.Context, tenantID string) (*TenantBindingMutationOutput, error)
	// ListTenantDeviceBindings returns paged tenant-device strategy bindings.
	ListTenantDeviceBindings(ctx context.Context, in ListBindingsInput) (*ListBindingsOutput, error)
	// SaveTenantDeviceBinding creates or updates one tenant-device strategy binding.
	SaveTenantDeviceBinding(ctx context.Context, in TenantDeviceBindingMutationInput) (*TenantDeviceBindingMutationOutput, error)
	// DeleteTenantDeviceBinding deletes one tenant-device strategy binding.
	DeleteTenantDeviceBinding(ctx context.Context, tenantID string, deviceID string) (*TenantDeviceBindingMutationOutput, error)
	// ResolveStrategy resolves the effective strategy for one tenant/device pair.
	ResolveStrategy(ctx context.Context, in ResolveStrategyInput) (*ResolveStrategyOutput, error)
	// ListAliases returns paged stream aliases.
	ListAliases(ctx context.Context, in ListAliasesInput) (*ListAliasesOutput, error)
	// GetAlias returns one stream alias by ID.
	GetAlias(ctx context.Context, id int64) (*AliasOutput, error)
	// CreateAlias creates one stream alias.
	CreateAlias(ctx context.Context, in AliasMutationInput) (int64, error)
	// UpdateAlias updates one stream alias.
	UpdateAlias(ctx context.Context, id int64, in AliasMutationInput) error
	// DeleteAlias deletes one stream alias.
	DeleteAlias(ctx context.Context, id int64) error
	// ListTenantWhites returns paged tenant whitelist entries.
	ListTenantWhites(ctx context.Context, in ListTenantWhitesInput) (*ListTenantWhitesOutput, error)
	// GetTenantWhite returns one tenant whitelist entry by natural key.
	GetTenantWhite(ctx context.Context, tenantID string, ip string) (*TenantWhiteOutput, error)
	// CreateTenantWhite creates one tenant whitelist entry.
	CreateTenantWhite(ctx context.Context, in TenantWhiteMutationInput) (*TenantWhiteMutationOutput, error)
	// UpdateTenantWhite updates one tenant whitelist entry.
	UpdateTenantWhite(ctx context.Context, tenantID string, ip string, in TenantWhiteMutationInput) (*TenantWhiteMutationOutput, error)
	// DeleteTenantWhite deletes one tenant whitelist entry.
	DeleteTenantWhite(ctx context.Context, tenantID string, ip string) (*TenantWhiteMutationOutput, error)
}

// Interface compliance assertion for the default media service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc bizctx.Service // bizCtxSvc reads current user and tenant metadata.
}

// New creates and returns a new media service instance.
func New() Service {
	return &serviceImpl{bizCtxSvc: bizctx.New()}
}
