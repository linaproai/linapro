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
// authentication provider registered by source plugins. The projection is
// intentionally public and static so the high-traffic login page does not read
// provider settings or expose SSO redirect rules.
func (s *serviceImpl) ListProviders(ctx context.Context) ([]ProviderEntryView, error) {
	checker := func(checkCtx context.Context, pluginID string) bool {
		if s == nil || s.pluginSvc == nil {
			return true
		}
		return s.pluginSvc.IsProviderEnabled(checkCtx, pluginID)
	}
	views, err := authprovider.ListViews(ctx, checker)
	if err != nil {
		return nil, err
	}
	out := make([]ProviderEntryView, 0, len(views))
	for _, v := range views {
		out = append(out, ProviderEntryView{
			ProviderID:   v.ProviderID,
			PluginID:     v.PluginID,
			Kind:         v.Kind.String(),
			Name:         v.Name,
			Description:  v.Description,
			Icon:         v.Icon,
			EntryURL:     v.EntryURL,
			DisplayOrder: v.DisplayOrder,
			Enabled:      v.Enabled,
		})
	}
	return out, nil
}
