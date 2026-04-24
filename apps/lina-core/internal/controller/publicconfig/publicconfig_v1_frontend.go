// This file handles the public frontend-config whitelist endpoint.

package publicconfig

import (
	"context"
	"strings"

	"lina-core/api/publicconfig/v1"
	hostconfig "lina-core/internal/service/config"
)

// Frontend returns the public-safe frontend display config whitelist.
func (c *ControllerV1) Frontend(ctx context.Context, _ *v1.FrontendReq) (res *v1.FrontendRes, err error) {
	cfg := c.configSvc.GetPublicFrontend(ctx)
	if cfg == nil {
		return &v1.FrontendRes{}, nil
	}
	return &v1.FrontendRes{
		App: v1.FrontendAppRes{
			Name:     c.localizePublicFrontendText(ctx, hostconfig.PublicFrontendSettingKeyAppName, "publicFrontend.app.name", cfg.App.Name),
			Logo:     cfg.App.Logo,
			LogoDark: cfg.App.LogoDark,
		},
		Auth: v1.FrontendAuthRes{
			PageTitle:     c.localizePublicFrontendText(ctx, hostconfig.PublicFrontendSettingKeyAuthPageTitle, "publicFrontend.auth.pageTitle", cfg.Auth.PageTitle),
			PageDesc:      c.localizePublicFrontendText(ctx, hostconfig.PublicFrontendSettingKeyAuthPageDesc, "publicFrontend.auth.pageDesc", cfg.Auth.PageDesc),
			LoginSubtitle: c.localizePublicFrontendText(ctx, hostconfig.PublicFrontendSettingKeyAuthLoginSubtitle, "publicFrontend.auth.loginSubtitle", cfg.Auth.LoginSubtitle),
			PanelLayout:   string(cfg.Auth.PanelLayout),
		},
		UI: v1.FrontendUIRes{
			ThemeMode:        cfg.UI.ThemeMode,
			Layout:           cfg.UI.Layout,
			WatermarkEnabled: cfg.UI.WatermarkEnabled,
			WatermarkContent: c.localizePublicFrontendText(ctx, hostconfig.PublicFrontendSettingKeyUIWatermarkContent, "publicFrontend.ui.watermarkContent", cfg.UI.WatermarkContent),
		},
		Cron: v1.FrontendCronRes{
			LogRetention: v1.FrontendCronLogRetentionRes{
				Mode:  string(cfg.Cron.LogRetention.Mode),
				Value: cfg.Cron.LogRetention.Value,
			},
			Shell: v1.FrontendCronShellRes{
				Enabled:        cfg.Cron.Shell.Enabled,
				Supported:      cfg.Cron.Shell.Supported,
				DisabledReason: cfg.Cron.Shell.DisabledReason,
			},
			Timezone: v1.FrontendCronTimezoneRes{
				Current: cfg.Cron.Timezone.Current,
			},
		},
	}, nil
}

// localizePublicFrontendText translates one public-frontend text field only when
// it still equals the built-in baseline value. Custom runtime values remain the
// source of truth until multi-language overrides are added for sys_config.
func (c *ControllerV1) localizePublicFrontendText(ctx context.Context, configKey string, messageKey string, current string) string {
	spec, ok := hostconfig.LookupPublicFrontendSettingSpec(configKey)
	if !ok {
		return current
	}
	if strings.TrimSpace(current) != "" && strings.TrimSpace(current) != strings.TrimSpace(spec.DefaultValue) {
		return current
	}
	return c.i18nSvc.Translate(ctx, messageKey, current)
}
