// This file handles the public frontend-config whitelist endpoint.

package publicconfig

import (
	"context"
	"strings"

	v1 "lina-core/api/publicconfig/v1"
	hostconfig "lina-core/internal/service/config"
)

// Frontend returns the public-safe frontend display config whitelist.
func (c *ControllerV1) Frontend(ctx context.Context, _ *v1.FrontendReq) (res *v1.FrontendRes, err error) {
	cfg, err := c.configSvc.GetPublicFrontend(ctx)
	if err != nil {
		return nil, err
	}
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
			PageTitle:             c.localizePublicFrontendText(ctx, hostconfig.PublicFrontendSettingKeyAuthPageTitle, "publicFrontend.auth.pageTitle", cfg.Auth.PageTitle),
			PageDesc:              c.localizePublicFrontendText(ctx, hostconfig.PublicFrontendSettingKeyAuthPageDesc, "publicFrontend.auth.pageDesc", cfg.Auth.PageDesc),
			LoginSubtitle:         c.localizePublicFrontendText(ctx, hostconfig.PublicFrontendSettingKeyAuthLoginSubtitle, "publicFrontend.auth.loginSubtitle", cfg.Auth.LoginSubtitle),
			PanelLayout:           v1.PanelLayout(cfg.Auth.PanelLayout),
			SloganImage:           cfg.Auth.SloganImage,
			ForgetPasswordEnabled: cfg.Auth.ForgetPasswordEnabled,
			RegisterEnabled:       cfg.Auth.RegisterEnabled,
			PrivacyPolicy:         c.localizePublicFrontendText(ctx, hostconfig.PublicFrontendSettingKeyAuthPrivacyPolicy, "publicFrontend.auth.privacyPolicy", cfg.Auth.PrivacyPolicy),
			TermsOfService:        c.localizePublicFrontendText(ctx, hostconfig.PublicFrontendSettingKeyAuthTermsOfService, "publicFrontend.auth.termsOfService", cfg.Auth.TermsOfService),
		},
		User: v1.FrontendUserRes{
			DefaultAvatar: cfg.User.DefaultAvatar,
		},
		UI: v1.FrontendUIRes{
			ThemeMode:        v1.ThemeMode(cfg.UI.ThemeMode),
			Layout:           v1.Layout(cfg.UI.Layout),
			WatermarkEnabled: cfg.UI.WatermarkEnabled,
			WatermarkContent: c.localizePublicFrontendText(ctx, hostconfig.PublicFrontendSettingKeyUIWatermarkContent, "publicFrontend.ui.watermarkContent", cfg.UI.WatermarkContent),
		},
		Cron: v1.FrontendCronRes{
			LogRetention: v1.FrontendCronLogRetentionRes{
				Mode:  v1.CronLogRetentionMode(cfg.Cron.LogRetention.Mode),
				Value: cfg.Cron.LogRetention.Value,
			},
			Shell: v1.FrontendCronShellRes{
				Enabled:           cfg.Cron.Shell.Enabled,
				Supported:         cfg.Cron.Shell.Supported,
				DisabledReason:    cfg.Cron.Shell.DisabledReason,
				DisabledReasonKey: cfg.Cron.Shell.DisabledReasonKey,
			},
			Timezone: v1.FrontendCronTimezoneRes{
				Current: cfg.Cron.Timezone.Current,
			},
		},
		Workspace: v1.FrontendWorkspaceRes{
			BasePath: cfg.Workspace.BasePath,
		},
	}, nil
}

// localizePublicFrontendText translates one public-frontend text field only
// when it still equals the built-in source text or the default-locale seed text.
// Custom runtime values remain the source of truth until multi-language
// overrides are added for sys_config.
func (c *ControllerV1) localizePublicFrontendText(ctx context.Context, configKey string, messageKey string, current string) string {
	spec, ok := hostconfig.LookupPublicFrontendSettingSpec(configKey)
	if !ok {
		return current
	}
	if c.i18nSvc == nil {
		return current
	}
	trimmedCurrent := strings.TrimSpace(current)
	if trimmedCurrent != "" &&
		trimmedCurrent != strings.TrimSpace(spec.DefaultValue) &&
		trimmedCurrent != c.defaultPublicFrontendText(messageKey, spec.DefaultValue) {
		return current
	}
	return c.i18nSvc.Translate(ctx, messageKey, current)
}

// defaultPublicFrontendText returns the default-locale public-frontend baseline
// used by SQL seed data. It deliberately uses a root context so the comparison
// does not depend on the request locale currently being projected.
func (c *ControllerV1) defaultPublicFrontendText(messageKey string, fallback string) string {
	return strings.TrimSpace(c.i18nSvc.Translate(context.Background(), messageKey, fallback))
}
