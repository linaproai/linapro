// This file implements the public /auth/providers endpoint that the default
// workbench login page reads to render third-party login entries discovered
// from source-plugin auth providers.

package auth

import (
	"context"

	v1 "lina-core/api/auth/v1"
	authsvc "lina-core/internal/service/auth"
)

// ListProviders returns enabled authentication provider entries discovered by
// the host auth-provider registry. Entries owned by disabled plugins are
// excluded so the login page reflects current workbench state without an
// extra plugin-status lookup.
func (c *ControllerV1) ListProviders(ctx context.Context, _ *v1.ListProvidersReq) (*v1.ListProvidersRes, error) {
	if c == nil || c.authSvc == nil {
		return &v1.ListProvidersRes{Providers: nil}, nil
	}
	views, err := c.authSvc.ListProviders(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListProvidersRes{Providers: toProviderEntities(views)}, nil
}

// toProviderEntities projects host-level provider views into API entities,
// preserving stable field names and zero-value defaults.
func toProviderEntities(views []authsvc.ProviderEntryView) []*v1.ProviderEntity {
	if len(views) == 0 {
		return nil
	}
	entities := make([]*v1.ProviderEntity, 0, len(views))
	for _, v := range views {
		entities = append(entities, &v1.ProviderEntity{
			ProviderID:             v.ProviderID,
			PluginID:               v.PluginID,
			Kind:                   v.Kind,
			Name:                   v.Name,
			Description:            v.Description,
			Icon:                   v.Icon,
			EntryURL:               v.EntryURL,
			BackendRedirectEnabled: v.BackendRedirectEnabled,
			BackendRedirectDefault: v.BackendRedirectDefault,
			BackendRedirectRules:   v.BackendRedirectRules,
			DisplayOrder:           v.DisplayOrder,
		})
	}
	return entities
}
