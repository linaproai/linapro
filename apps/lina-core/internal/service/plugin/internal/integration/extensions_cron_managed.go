// This file collects plugin-owned cron definitions into stable projection
// records so the host can surface them in scheduled-job management.

package integration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

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
	if handler == nil {
		return gerror.New("插件定时任务处理器不能为空")
	}

	trimmedPattern := strings.TrimSpace(pattern)
	trimmedName := strings.TrimSpace(name)
	if trimmedPattern == "" {
		return gerror.New("插件定时任务表达式不能为空")
	}
	if trimmedName == "" {
		return gerror.New("插件定时任务名称不能为空")
	}

	c.items = append(c.items, ManagedCronJob{
		PluginID:       strings.TrimSpace(c.pluginID),
		Name:           trimmedName,
		DisplayName:    trimmedName,
		Description:    fmt.Sprintf("插件 %s 注册的内置定时任务。", strings.TrimSpace(c.pluginID)),
		Pattern:        trimmedPattern,
		Timezone:       "Asia/Shanghai",
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
// source plugins without registering them into gcron.
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
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok || sourcePlugin == nil {
			continue
		}

		collector := &managedCronCollector{
			pluginID: manifest.ID,
			items:    make([]ManagedCronJob, 0),
		}
		for _, registration := range sourcePlugin.GetCronRegistrars() {
			if registration == nil || registration.Handler == nil {
				continue
			}
			if err = registration.Handler(ctx, collector); err != nil {
				return nil, err
			}
		}
		result = append(result, collector.items...)
	}
	return result, nil
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
