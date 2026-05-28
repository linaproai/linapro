// This file implements ListProviders by consulting the process-level
// authprovider registry and filtering out providers whose owning plugin is
// currently disabled. The host /auth/providers endpoint and source-plugin
// AuthService.LoginByExternal handoff both consume this projection.

package auth

import (
	"context"

	"lina-core/pkg/plugin/capability/authprovider"
)

// ListProviders returns ProviderEntryView entries for every enabled
// authentication provider registered by source plugins. Each entry's
// runtime fields (BackendRedirectEnabled, BackendRedirectDefault,
// BackendRedirectRules) reflect the owning plugin's live settings so
// admin updates take effect without a server restart.
func (s *serviceImpl) ListProviders(ctx context.Context) ([]ProviderEntryView, error) {
	checker := func(checkCtx context.Context, pluginID string) bool {
		if s == nil || s.pluginSvc == nil {
			return true
		}
		return s.pluginSvc.IsEnabled(checkCtx, pluginID)
	}
	views, err := authprovider.ListViews(ctx, checker)
	if err != nil {
		return nil, err
	}
	out := make([]ProviderEntryView, 0, len(views))
	for _, v := range views {
		out = append(out, ProviderEntryView{
			ProviderID:             v.ProviderID,
			PluginID:               v.PluginID,
			Kind:                   v.Kind.String(),
			Name:                   v.Name,
			Description:            v.Description,
			Icon:                   v.Icon,
			EntryURL:               v.EntryURL,
			BackendRedirectEnabled: v.BackendRedirectEnabled,
			BackendRedirectDefault: v.BackendRedirectDefault,
			BackendRedirectRules:   v.BackendRedirectRules,
			DisplayOrder:           v.DisplayOrder,
			Enabled:                v.Enabled,
		})
	}
	return out, nil
}
