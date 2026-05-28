// This file exposes the source-plugin host service directory adapters.

package pluginhostservices

import (
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/contract"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	capabilitytenantcap "lina-core/pkg/plugin/capability/tenantcap"
)

// APIDoc returns the host API-documentation localization adapter.
func (s *directory) APIDoc() contract.APIDocService {
	if s == nil {
		return nil
	}
	return s.apiDoc
}

// Auth returns the host tenant-auth adapter.
func (s *directory) Auth() contract.AuthService {
	if s == nil {
		return nil
	}
	return s.auth
}

// BizCtx returns the host business-context adapter.
func (s *directory) BizCtx() contract.BizCtxService {
	if s == nil {
		return nil
	}
	return s.bizCtx
}

// Cache returns nil for the unscoped base directory because cache operations
// require a plugin-bound service view.
func (s *directory) Cache() contract.CacheService {
	return nil
}

// Config returns nil for the unscoped base directory because config reads
// require a plugin-bound service view.
func (s *directory) Config() contract.ConfigService {
	return nil
}

// HostConfig returns the whitelisted public host config adapter.
func (s *directory) HostConfig() contract.HostConfigService {
	if s == nil {
		return nil
	}
	return s.hostConfig
}

// I18n returns the host runtime translation adapter.
func (s *directory) I18n() contract.I18nService {
	if s == nil {
		return nil
	}
	return s.i18n
}

// Manifest returns nil for the unscoped base directory because manifest reads
// require a plugin-bound service view.
func (s *directory) Manifest() contract.ManifestService {
	return nil
}

// Notify returns the host notification adapter.
func (s *directory) Notify() contract.NotifyService {
	if s == nil {
		return nil
	}
	return s.notify
}

// Org returns the organization capability service.
func (s *directory) Org() capabilityorgcap.Service {
	if s == nil {
		return nil
	}
	return s.org
}

// PluginLifecycle returns the host plugin lifecycle orchestration adapter.
func (s *directory) PluginLifecycle() contract.PluginLifecycleService {
	if s == nil {
		return nil
	}
	return s.pluginLife
}

// PluginSettings returns the host namespaced key-value settings adapter.
func (s *directory) PluginSettings() contract.PluginSettingsService {
	if s == nil {
		return nil
	}
	return s.pluginSettings
}

// PluginState returns the host plugin enablement adapter.
func (s *directory) PluginState() contract.PluginStateService {
	if s == nil {
		return nil
	}
	return s.pluginState
}

// Route returns the host dynamic-route metadata adapter.
func (s *directory) Route() contract.RouteService {
	if s == nil {
		return nil
	}
	return s.route
}

// Session returns the host online-session adapter.
func (s *directory) Session() contract.SessionService {
	if s == nil {
		return nil
	}
	return s.session
}

// TenantFilter returns the host tenant-filter adapter.
func (s *directory) TenantFilter() contract.TenantFilterService {
	if s == nil {
		return nil
	}
	return s.tenantFilter
}

// Tenant returns the tenant capability service.
func (s *directory) Tenant() capabilitytenantcap.Service {
	if s == nil {
		return nil
	}
	return s.tenant
}

// ForPlugin returns a plugin-bound host service view.
func (s *directory) ForPlugin(pluginID string) capability.Services {
	if s == nil {
		return nil
	}
	return &scopedDirectory{base: s, pluginID: pluginID}
}

// APIDoc returns the delegated API-documentation localization adapter.
func (s *scopedDirectory) APIDoc() contract.APIDocService {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.APIDoc()
}

// Auth returns the delegated tenant-auth adapter.
func (s *scopedDirectory) Auth() contract.AuthService {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Auth()
}

// BizCtx returns the delegated business-context adapter.
func (s *scopedDirectory) BizCtx() contract.BizCtxService {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.BizCtx()
}

// Cache returns the plugin-scoped host cache adapter.
func (s *scopedDirectory) Cache() contract.CacheService {
	if s == nil || s.base == nil {
		return nil
	}
	return newCacheAdapter(s.base.cache, s.base.bizCtx, s.pluginID)
}

// Config returns the plugin-scoped static configuration adapter.
func (s *scopedDirectory) Config() contract.ConfigService {
	if s == nil || s.base == nil {
		return nil
	}
	if s.base.config == nil {
		return nil
	}
	return s.base.config.ForPlugin(s.pluginID)
}

// HostConfig returns the delegated public host config adapter.
func (s *scopedDirectory) HostConfig() contract.HostConfigService {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.HostConfig()
}

// I18n returns the delegated runtime translation adapter.
func (s *scopedDirectory) I18n() contract.I18nService {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.I18n()
}

// Manifest returns the plugin-scoped manifest resource adapter.
func (s *scopedDirectory) Manifest() contract.ManifestService {
	if s == nil || s.base == nil {
		return nil
	}
	if s.base.manifest == nil {
		return nil
	}
	return s.base.manifest.ForPlugin(s.pluginID)
}

// Notify returns the delegated notification adapter.
func (s *scopedDirectory) Notify() contract.NotifyService {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Notify()
}

// Org returns the delegated organization capability service.
func (s *scopedDirectory) Org() capabilityorgcap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Org()
}

// PluginLifecycle returns the delegated plugin lifecycle orchestration adapter.
func (s *scopedDirectory) PluginLifecycle() contract.PluginLifecycleService {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.PluginLifecycle()
}

// PluginSettings returns the delegated plugin settings adapter.
func (s *scopedDirectory) PluginSettings() contract.PluginSettingsService {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.PluginSettings()
}

// PluginState returns the delegated plugin enablement adapter.
func (s *scopedDirectory) PluginState() contract.PluginStateService {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.PluginState()
}

// Route returns the delegated dynamic-route metadata adapter.
func (s *scopedDirectory) Route() contract.RouteService {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Route()
}

// Session returns the delegated online-session adapter.
func (s *scopedDirectory) Session() contract.SessionService {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Session()
}

// TenantFilter returns the delegated tenant-filter adapter.
func (s *scopedDirectory) TenantFilter() contract.TenantFilterService {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.TenantFilter()
}

// Tenant returns the delegated tenant capability service.
func (s *scopedDirectory) Tenant() capabilitytenantcap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Tenant()
}
