// This file exposes the source-plugin host service directory adapters.

package pluginhostservices

import "lina-core/pkg/pluginservice/contract"

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

// Config returns the host static configuration adapter.
func (s *directory) Config() contract.ConfigService {
	if s == nil {
		return nil
	}
	return s.config
}

// I18n returns the host runtime translation adapter.
func (s *directory) I18n() contract.I18nService {
	if s == nil {
		return nil
	}
	return s.i18n
}

// Notify returns the host notification adapter.
func (s *directory) Notify() contract.NotifyService {
	if s == nil {
		return nil
	}
	return s.notify
}

// PluginLifecycle returns the host plugin lifecycle orchestration adapter.
func (s *directory) PluginLifecycle() contract.PluginLifecycleService {
	if s == nil {
		return nil
	}
	return s.pluginLife
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
