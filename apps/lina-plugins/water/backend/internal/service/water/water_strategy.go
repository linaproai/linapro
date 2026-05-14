// This file resolves media strategies and parses watermark strategy YAML.

package water

import (
	"context"
	"strings"

	"gopkg.in/yaml.v3"

	"lina-core/pkg/bizerr"
	"lina-plugin-water/backend/internal/dao"
	"lina-plugin-water/backend/internal/model/do"
	entitymodel "lina-plugin-water/backend/internal/model/entity"
)

// strategyEntity reuses the plugin-local generated media strategy entity.
type strategyEntity = entitymodel.MediaStrategy

// strategyYAML is a tolerant projection of media strategy YAML.
type strategyYAML struct {
	Watermark       *watermarkConfig `json:"watermark" yaml:"watermark"` // Watermark is the nested Lina strategy node.
	watermarkConfig `yaml:",inline"` // Inline fields keep hotgo root-level strategy compatibility.
}

// ResolveStrategy resolves one effective media strategy for watermark processing.
func ResolveStrategy(ctx context.Context, tenantID string, deviceID string) (*resolvedStrategy, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	tenantID = strings.TrimSpace(tenantID)
	deviceID = strings.TrimSpace(deviceID)

	if tenantID != "" && deviceID != "" {
		strategy, err := strategyFromTenantDeviceBinding(ctx, tenantID, deviceID)
		if err != nil {
			return nil, err
		}
		if strategy != nil {
			return buildResolvedStrategy(StrategySourceTenantDevice, strategy), nil
		}
	}
	if deviceID != "" {
		strategy, err := strategyFromDeviceBinding(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		if strategy != nil {
			return buildResolvedStrategy(StrategySourceDevice, strategy), nil
		}
	}
	if tenantID != "" {
		strategy, err := strategyFromTenantBinding(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		if strategy != nil {
			return buildResolvedStrategy(StrategySourceTenant, strategy), nil
		}
	}
	strategy, err := globalStrategy(ctx)
	if err != nil {
		return nil, err
	}
	if strategy != nil {
		return buildResolvedStrategy(StrategySourceGlobal, strategy), nil
	}
	return buildResolvedStrategy(StrategySourceNone, nil), nil
}

// parseWatermarkStrategy parses watermark configuration from strategy YAML.
func parseWatermarkStrategy(strategyBody string) (*watermarkConfig, error) {
	body := strings.TrimSpace(strategyBody)
	if body == "" {
		return nil, nil
	}
	var parsed strategyYAML
	if err := yaml.Unmarshal([]byte(body), &parsed); err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyParseFailed)
	}

	cfg := parsed.Watermark
	if cfg == nil && hasRootWatermarkConfig(parsed.watermarkConfig) {
		cfg = &parsed.watermarkConfig
	}
	if cfg == nil {
		return nil, nil
	}
	normalized := normalizeWatermarkConfig(*cfg)
	return &normalized, nil
}

// validateMediaTablesReady verifies the media-owned shared strategy tables exist.
func validateMediaTablesReady(ctx context.Context) error {
	tableNames := []string{
		dao.MediaStrategy.Table(),
		dao.MediaStrategyDevice.Table(),
		dao.MediaStrategyDeviceTenant.Table(),
		dao.MediaStrategyTenant.Table(),
	}
	for _, tableName := range tableNames {
		fields, err := dao.MediaStrategy.DB().TableFields(ctx, tableName)
		if err != nil {
			return bizerr.WrapCode(err, CodeWaterMediaTableCheckFailed)
		}
		if len(fields) == 0 {
			return bizerr.NewCode(CodeWaterMediaTableNotInstalled)
		}
	}
	return nil
}

// strategyFromTenantDeviceBinding returns the enabled strategy bound to a tenant-device pair.
func strategyFromTenantDeviceBinding(ctx context.Context, tenantID string, deviceID string) (*strategyEntity, error) {
	var binding *entitymodel.MediaStrategyDeviceTenant
	err := dao.MediaStrategyDeviceTenant.Ctx(ctx).
		Where(do.MediaStrategyDeviceTenant{TenantId: tenantID, DeviceId: deviceID}).
		Scan(&binding)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyQueryFailed)
	}
	if binding == nil {
		return nil, nil
	}
	return enabledStrategyByID(ctx, binding.StrategyId)
}

// strategyFromDeviceBinding returns the enabled strategy bound to a device.
func strategyFromDeviceBinding(ctx context.Context, deviceID string) (*strategyEntity, error) {
	var binding *entitymodel.MediaStrategyDevice
	err := dao.MediaStrategyDevice.Ctx(ctx).
		Where(do.MediaStrategyDevice{DeviceId: deviceID}).
		Scan(&binding)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyQueryFailed)
	}
	if binding == nil {
		return nil, nil
	}
	return enabledStrategyByID(ctx, binding.StrategyId)
}

// strategyFromTenantBinding returns the enabled strategy bound to a tenant.
func strategyFromTenantBinding(ctx context.Context, tenantID string) (*strategyEntity, error) {
	var binding *entitymodel.MediaStrategyTenant
	err := dao.MediaStrategyTenant.Ctx(ctx).
		Where(do.MediaStrategyTenant{TenantId: tenantID}).
		Scan(&binding)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyQueryFailed)
	}
	if binding == nil {
		return nil, nil
	}
	return enabledStrategyByID(ctx, binding.StrategyId)
}

// globalStrategy returns the enabled global strategy.
func globalStrategy(ctx context.Context) (*strategyEntity, error) {
	var strategy *strategyEntity
	err := dao.MediaStrategy.Ctx(ctx).
		Where(dao.MediaStrategy.Columns().Global, switchOn).
		Where(dao.MediaStrategy.Columns().Enable, switchOn).
		OrderDesc(dao.MediaStrategy.Columns().Id).
		Scan(&strategy)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyQueryFailed)
	}
	return strategy, nil
}

// enabledStrategyByID returns one enabled strategy by ID.
func enabledStrategyByID(ctx context.Context, id int64) (*strategyEntity, error) {
	var strategy *strategyEntity
	err := dao.MediaStrategy.Ctx(ctx).
		Where(dao.MediaStrategy.Columns().Enable, switchOn).
		Where(do.MediaStrategy{Id: id}).
		Scan(&strategy)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyQueryFailed)
	}
	return strategy, nil
}

// buildResolvedStrategy converts a strategy source and entity into service output.
func buildResolvedStrategy(source StrategySource, strategy *strategyEntity) *resolvedStrategy {
	out := &resolvedStrategy{
		Matched:     strategy != nil,
		Source:      source,
		SourceLabel: strategySourceLabel(source),
	}
	if strategy != nil {
		out.StrategyId = strategy.Id
		out.StrategyName = strategy.Name
		out.Strategy = strategy.Strategy
	}
	return out
}

// strategySourceLabel returns the Chinese label for one strategy source.
func strategySourceLabel(source StrategySource) string {
	switch source {
	case StrategySourceTenantDevice:
		return "租户设备策略"
	case StrategySourceDevice:
		return "设备策略"
	case StrategySourceTenant:
		return "租户策略"
	case StrategySourceGlobal:
		return "全局策略"
	default:
		return "未匹配"
	}
}

// hasRootWatermarkConfig reports whether the root YAML object looks like a hotgo watermark config.
func hasRootWatermarkConfig(cfg watermarkConfig) bool {
	return cfg.Enabled ||
		strings.TrimSpace(cfg.Text) != "" ||
		strings.TrimSpace(cfg.Image) != "" ||
		strings.TrimSpace(cfg.Base64) != "" ||
		cfg.FontSize > 0 ||
		cfg.Opacity > 0 ||
		cfg.PosX > 0 ||
		cfg.PosY > 0 ||
		cfg.Align != ""
}

// normalizeWatermarkConfig fills defaults and normalizes bounded fields.
func normalizeWatermarkConfig(cfg watermarkConfig) watermarkConfig {
	cfg.Text = strings.TrimSpace(cfg.Text)
	cfg.Font = strings.TrimSpace(cfg.Font)
	cfg.Color = strings.TrimSpace(cfg.Color)
	cfg.Image = strings.TrimSpace(cfg.Image)
	cfg.Base64 = strings.TrimSpace(cfg.Base64)
	cfg.Align = watermarkAlignment(strings.TrimSpace(string(cfg.Align)))
	if cfg.FontSize <= 0 {
		cfg.FontSize = defaultFontSize
	}
	if cfg.Color == "" {
		cfg.Color = "#ffffff"
	}
	if cfg.Opacity <= 0 {
		cfg.Opacity = defaultWatermarkOpacity
	}
	if cfg.Opacity > 1 {
		cfg.Opacity = 1
	}
	return cfg
}

// watermarkAlignment accepts both hotgo numeric and Lina named alignment values.
type watermarkAlignment string

// UnmarshalYAML decodes alignment from YAML strings or integers.
func (a *watermarkAlignment) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		*a = watermarkAlignment(strings.TrimSpace(value.Value))
		return nil
	default:
		*a = ""
		return nil
	}
}
