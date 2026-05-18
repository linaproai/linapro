// Package sourceupgrade implements source-plugin upgrade discovery and explicit
// runtime upgrade execution for the host plugin domain.
package sourceupgrade

import (
	"context"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/runtime"
	sourceupgradecontract "lina-core/pkg/sourceupgrade/contract"
)

type (
	// SourceUpgradeStatus aliases the stable source-plugin upgrade status contract.
	SourceUpgradeStatus = sourceupgradecontract.SourcePluginStatus

	// SourceUpgradeResult aliases the stable explicit source-plugin upgrade result contract.
	SourceUpgradeResult = sourceupgradecontract.SourcePluginUpgradeResult
)

// Service defines the host-side source-plugin upgrade governance contract.
type Service interface {
	// ListSourceUpgradeStatuses scans source manifests and returns one
	// effective-versus-discovered upgrade-status item per source plugin.
	ListSourceUpgradeStatuses(ctx context.Context) ([]*SourceUpgradeStatus, error)
	// UpgradeSourcePlugin applies one explicit source-plugin upgrade from the
	// current effective version to the newer discovered source version.
	UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourceUpgradeResult, error)
	// ValidateSourcePluginUpgradeReadiness scans source-plugin version drift
	// without failing on pending upgrades.
	ValidateSourcePluginUpgradeReadiness(ctx context.Context) error
}

// DependencyValidator validates source-plugin upgrade candidates before the
// upgrade service runs SQL, menu sync, or release switching.
type DependencyValidator interface {
	// ValidateSourcePluginUpgradeCandidate verifies candidate dependencies and
	// reverse-dependency version safety for one source plugin upgrade.
	ValidateSourcePluginUpgradeCandidate(ctx context.Context, manifest *catalog.Manifest) error
}

// Ensure serviceImpl satisfies Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest discovery, registry, and release governance.
	catalogSvc catalog.Service
	// lifecycleSvc provides install/uninstall lifecycle orchestration.
	lifecycleSvc lifecycle.Service
	// runtimeSvc provides dynamic plugin reconciliation and route dispatch.
	runtimeSvc runtime.Service
	// integrationSvc provides host extension, menu, hook, and resource integration.
	integrationSvc integration.Service
	// i18nSvc localizes operator-facing result messages.
	i18nSvc sourceUpgradeI18nService
	// dependencyValidator checks candidate release dependency constraints before upgrade side effects.
	dependencyValidator DependencyValidator
}

// sourceUpgradeI18nService defines the narrow i18n capability needed by source upgrade.
type sourceUpgradeI18nService interface {
	// Translate returns one runtime translation key with caller-provided fallback text.
	Translate(ctx context.Context, key string, fallback string) string
}

// New creates and returns a new source-plugin upgrade governance service.
func New(
	catalogSvc catalog.Service,
	lifecycleSvc lifecycle.Service,
	runtimeSvc runtime.Service,
	integrationSvc integration.Service,
	i18nSvc sourceUpgradeI18nService,
	dependencyValidator DependencyValidator,
) Service {
	return &serviceImpl{
		catalogSvc:          catalogSvc,
		lifecycleSvc:        lifecycleSvc,
		runtimeSvc:          runtimeSvc,
		integrationSvc:      integrationSvc,
		i18nSvc:             i18nSvc,
		dependencyValidator: dependencyValidator,
	}
}
