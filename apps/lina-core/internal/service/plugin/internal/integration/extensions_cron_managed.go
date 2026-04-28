// This file collects plugin-owned cron definitions into stable projection
// records so the host can surface them in scheduled-job management.

package integration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/jobmeta"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginhost"
)

const (
	managedCronDefaultTimeout = 5 * time.Minute
)

// managedCronCollector captures plugin-owned cron registrations instead of
// registering them directly into gcron.
type managedCronCollector struct {
	pluginID string
	items    []ManagedCronJob
}

// Ensure managedCronCollector satisfies the published registrar contract.
var _ pluginhost.CronRegistrar = (*managedCronCollector)(nil)

// Add records one plugin-owned cron job definition.
func (c *managedCronCollector) Add(
	ctx context.Context,
	pattern string,
	name string,
	handler pluginhost.CronJobHandler,
) error {
	return c.AddWithMetadata(
		ctx,
		pattern,
		name,
		name,
		fmt.Sprintf("Built-in scheduled job registered by plugin %s.", strings.TrimSpace(c.pluginID)),
		handler,
	)
}

// AddWithMetadata records one plugin-owned cron job definition with display metadata.
func (c *managedCronCollector) AddWithMetadata(
	ctx context.Context,
	pattern string,
	name string,
	displayName string,
	description string,
	handler pluginhost.CronJobHandler,
) error {
	if handler == nil {
		return gerror.New("plugin cron job handler cannot be nil")
	}

	trimmedPattern := strings.TrimSpace(pattern)
	trimmedName := strings.TrimSpace(name)
	trimmedDisplayName := strings.TrimSpace(displayName)
	trimmedDescription := strings.TrimSpace(description)
	if trimmedPattern == "" {
		return gerror.New("plugin cron job expression cannot be empty")
	}
	if trimmedName == "" {
		return gerror.New("plugin cron job name cannot be empty")
	}
	if trimmedDisplayName == "" {
		trimmedDisplayName = trimmedName
	}
	if trimmedDescription == "" {
		trimmedDescription = fmt.Sprintf("Built-in scheduled job registered by plugin %s.", strings.TrimSpace(c.pluginID))
	}

	c.items = append(c.items, ManagedCronJob{
		PluginID:       strings.TrimSpace(c.pluginID),
		Name:           trimmedName,
		DisplayName:    trimmedDisplayName,
		Description:    trimmedDescription,
		Pattern:        trimmedPattern,
		Timezone:       pluginbridge.DefaultCronContractTimezone,
		Scope:          "", // Legacy RegisterCron callbacks do not expose scope metadata.
		Concurrency:    "",
		MaxConcurrency: 1,
		Timeout:        managedCronDefaultTimeout,
		Handler:        handler,
	})
	return nil
}

// IsPrimaryNode reports a stable true value while collecting definitions so
// source plugins do not accidentally hide jobs from the unified registry view.
func (c *managedCronCollector) IsPrimaryNode() bool {
	return true
}

// collectManagedCronJobs gathers plugin-owned cron definitions from matching
// source and dynamic plugins without registering them into gcron.
func (s *serviceImpl) collectManagedCronJobs(
	ctx context.Context,
	pluginID string,
) ([]ManagedCronJob, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}

	result := make([]ManagedCronJob, 0)
	trimmedPluginID := strings.TrimSpace(pluginID)
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		if trimmedPluginID != "" && manifest.ID != trimmedPluginID {
			continue
		}

		sourceItems, err := s.collectSourceManagedCronJobs(ctx, manifest)
		if err != nil {
			return nil, err
		}
		result = append(result, sourceItems...)

		dynamicItems, err := s.collectDynamicManagedCronJobs(ctx, manifest)
		if err != nil {
			return nil, err
		}
		result = append(result, dynamicItems...)
	}
	return result, nil
}

// collectSourceManagedCronJobs gathers source-plugin managed cron registrations
// for one manifest.
func (s *serviceImpl) collectSourceManagedCronJobs(
	ctx context.Context,
	manifest *catalog.Manifest,
) ([]ManagedCronJob, error) {
	if manifest == nil {
		return nil, nil
	}
	sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
	if !ok || sourcePlugin == nil {
		return nil, nil
	}

	collector := &managedCronCollector{
		pluginID: manifest.ID,
		items:    make([]ManagedCronJob, 0),
	}
	for _, registration := range sourcePlugin.GetCronRegistrars() {
		if registration == nil || registration.Handler == nil {
			continue
		}
		if err := registration.Handler(ctx, collector); err != nil {
			return nil, err
		}
	}
	return collector.items, nil
}

// collectDynamicManagedCronJobs gathers dynamic-plugin cron declarations from
// the runtime registration entry point and binds them to the shared executor.
func (s *serviceImpl) collectDynamicManagedCronJobs(
	ctx context.Context,
	manifest *catalog.Manifest,
) ([]ManagedCronJob, error) {
	if manifest == nil {
		return nil, nil
	}
	// Only runtime-loaded dynamic plugins expose cron contracts through the Wasm
	// discovery entry point. Source plugins are handled by the callback-based
	// collector above and must not be routed through dynamic discovery.
	if catalog.NormalizeType(manifest.Type) != catalog.TypeDynamic {
		return nil, nil
	}
	if !manifestDeclaresCronHostService(manifest) {
		return nil, nil
	}
	if s.dynamicCronExecutor == nil {
		return nil, gerror.Newf("dynamic plugin cron executor is not injected: %s", manifest.ID)
	}

	contracts, err := s.dynamicCronExecutor.DiscoverCronContracts(ctx, manifest)
	if err != nil {
		return nil, err
	}
	if len(contracts) == 0 {
		return nil, nil
	}

	items := make([]ManagedCronJob, 0, len(contracts))
	for _, contract := range contracts {
		if contract == nil {
			continue
		}
		contractSnapshot := *contract
		manifestSnapshot := manifest
		items = append(items, ManagedCronJob{
			PluginID:       strings.TrimSpace(manifest.ID),
			Name:           strings.TrimSpace(contractSnapshot.Name),
			DisplayName:    strings.TrimSpace(contractSnapshot.DisplayName),
			Description:    strings.TrimSpace(contractSnapshot.Description),
			Pattern:        strings.TrimSpace(contractSnapshot.Pattern),
			Timezone:       strings.TrimSpace(contractSnapshot.Timezone),
			Scope:          jobmeta.NormalizeJobScope(contractSnapshot.Scope.String()),
			Concurrency:    jobmeta.NormalizeJobConcurrency(contractSnapshot.Concurrency.String()),
			MaxConcurrency: contractSnapshot.MaxConcurrency,
			Timeout:        time.Duration(contractSnapshot.TimeoutSeconds) * time.Second,
			Handler: func(ctx context.Context) error {
				return s.dynamicCronExecutor.ExecuteDeclaredCronJob(ctx, manifestSnapshot, &contractSnapshot)
			},
		})
	}
	return items, nil
}

// manifestDeclaresCronHostService reports whether the manifest explicitly
// authorizes the dedicated cron host service for dynamic cron discovery.
func manifestDeclaresCronHostService(manifest *catalog.Manifest) bool {
	if manifest == nil {
		return false
	}
	for _, service := range manifest.HostServices {
		if service == nil {
			continue
		}
		if strings.TrimSpace(service.Service) == pluginbridge.HostServiceCron {
			return true
		}
	}
	return false
}

// ListManagedCronJobs returns all plugin-owned cron definitions for
// projection into the unified scheduled-job table.
func (s *serviceImpl) ListManagedCronJobs(ctx context.Context) ([]ManagedCronJob, error) {
	return s.collectManagedCronJobs(ctx, "")
}

// ListManagedCronJobsByPlugin returns cron definitions owned by one plugin.
func (s *serviceImpl) ListManagedCronJobsByPlugin(ctx context.Context, pluginID string) ([]ManagedCronJob, error) {
	return s.collectManagedCronJobs(ctx, pluginID)
}
