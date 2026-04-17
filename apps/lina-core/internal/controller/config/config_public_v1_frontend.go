// This file handles the public frontend-config whitelist endpoint.

package config

import (
	"context"

	v1 "lina-core/api/publicconfig/v1"
)

// Frontend returns the public-safe frontend display config whitelist.
func (c *PublicControllerV1) Frontend(ctx context.Context, _ *v1.FrontendReq) (res *v1.FrontendRes, err error) {
	cfg := c.configSvc.GetPublicFrontend(ctx)
	if cfg == nil {
		return &v1.FrontendRes{}, nil
	}
	return &v1.FrontendRes{
		App: v1.FrontendAppRes{
			Name:     cfg.App.Name,
			Logo:     cfg.App.Logo,
			LogoDark: cfg.App.LogoDark,
		},
		Auth: v1.FrontendAuthRes{
			PageTitle:     cfg.Auth.PageTitle,
			PageDesc:      cfg.Auth.PageDesc,
			LoginSubtitle: cfg.Auth.LoginSubtitle,
		},
		UI: v1.FrontendUIRes{
			ThemeMode:        cfg.UI.ThemeMode,
			Layout:           cfg.UI.Layout,
			WatermarkEnabled: cfg.UI.WatermarkEnabled,
			WatermarkContent: cfg.UI.WatermarkContent,
		},
	}, nil
}
