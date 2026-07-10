// This file exposes the source-plugin host service directory adapters.

package capabilityhost

import (
	filesvc "lina-core/internal/service/file"
	"lina-core/internal/service/notify"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	capabilitydictcap "lina-core/pkg/plugin/capability/dictcap"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	capabilityplugincap "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/storagecap"
	capabilitytenantcap "lina-core/pkg/plugin/capability/tenantcap"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
)

// APIDoc returns the host API-documentation localization adapter.
func (s *directory) APIDoc() apidoccap.Service {
	if s == nil {
		return nil
	}
	return s.apiDoc
}

// Auth returns the host authentication and authorization namespace.
func (s *directory) Auth() authcap.Service {
	if s == nil {
		return nil
	}
	return s.auth
}

// Users returns the user-domain ordinary capability service.
func (s *directory) Users() capabilityusercap.Service {
	if s == nil {
		return nil
	}
	return s.users
}

// BizCtx returns the host business-context adapter.
func (s *directory) BizCtx() bizctxcap.Service {
	if s == nil {
		return nil
	}
	return s.bizCtx
}

// Cache returns nil for the unscoped base directory because cache operations
// require a plugin-bound service view.
func (s *directory) Cache() cachecap.Service {
	return nil
}

// Dict returns the dictionary-domain ordinary capability service.
func (s *directory) Dict() capabilitydictcap.Service {
	if s == nil {
		return nil
	}
	return s.dict
}

// Files returns the file-domain ordinary capability service.
func (s *directory) Files() capabilityfilecap.Service {
	if s == nil {
		return nil
	}
	return s.files
}

// HostConfig returns the host config adapter.
func (s *directory) HostConfig() hostconfigcap.Service {
	if s == nil {
		return nil
	}
	return s.hostConfig
}

// I18n returns the host runtime translation adapter.
func (s *directory) I18n() i18ncap.Service {
	if s == nil {
		return nil
	}
	return s.i18n
}

// Jobs returns the scheduled-job domain ordinary capability service.
func (s *directory) Jobs() capabilityjobcap.Service {
	if s == nil {
		return nil
	}
	return s.jobs
}

// Lock returns nil for the unscoped base directory because lock operations
// require a plugin-bound service view.
func (s *directory) Lock() lockcap.Service {
	return nil
}

// Manifest returns nil for the unscoped base directory because manifest reads
// require a plugin-bound service view.
func (s *directory) Manifest() manifestcap.Service {
	return nil
}

// Notifications returns the notification-domain ordinary capability service.
func (s *directory) Notifications() capabilitynotifycap.Service {
	if s == nil {
		return nil
	}
	return s.notifications
}

// Org returns the organization capability service.
func (s *directory) Org() capabilityorgcap.Service {
	if s == nil {
		return nil
	}
	return s.org
}

// Plugins returns the plugin-governance ordinary capability service.
func (s *directory) Plugins() capabilityplugincap.Service {
	if s == nil {
		return nil
	}
	return s.plugins
}

// Route returns the host dynamic-route metadata adapter.
func (s *directory) Route() routecap.Service {
	if s == nil {
		return nil
	}
	return s.route
}

// Sessions returns the online-session domain ordinary capability service.
func (s *directory) Sessions() capabilitysessioncap.Service {
	if s == nil {
		return nil
	}
	return s.sessions
}

// Storage returns nil for the unscoped base directory because storage
// operations require a plugin-bound service view.
func (s *directory) Storage() storagecap.Service {
	return nil
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
func (s *scopedDirectory) APIDoc() apidoccap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.APIDoc()
}

// Auth returns the delegated authentication and authorization namespace.
func (s *scopedDirectory) Auth() authcap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Auth()
}

// Users returns the delegated user-domain ordinary capability service.
func (s *scopedDirectory) Users() capabilityusercap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Users()
}

// BizCtx returns the delegated business-context adapter.
func (s *scopedDirectory) BizCtx() bizctxcap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.BizCtx()
}

// Cache returns the plugin-scoped host cache adapter.
func (s *scopedDirectory) Cache() cachecap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return newCacheAdapter(s.base.cache, s.base.bizCtx, s.pluginID)
}

// Dict returns the delegated dictionary-domain ordinary capability service.
func (s *scopedDirectory) Dict() capabilitydictcap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Dict()
}

// Files returns the plugin-scoped file-domain ordinary capability service.
func (s *scopedDirectory) Files() capabilityfilecap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	files := s.base.Files()
	if scoped, ok := files.(filesvc.CapabilityService); ok {
		return scoped.WithStorage(s.Storage())
	}
	return files
}

// HostConfig returns the delegated host config adapter.
func (s *scopedDirectory) HostConfig() hostconfigcap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.HostConfig()
}

// I18n returns the delegated runtime translation adapter.
func (s *scopedDirectory) I18n() i18ncap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.I18n()
}

// Jobs returns the delegated scheduled-job domain ordinary capability service.
func (s *scopedDirectory) Jobs() capabilityjobcap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Jobs()
}

// Lock returns the plugin-scoped host lock adapter.
func (s *scopedDirectory) Lock() lockcap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return newLockAdapter(s.base.lock, s.base.bizCtx, s.pluginID)
}

// Manifest returns the plugin-scoped manifest resource adapter.
func (s *scopedDirectory) Manifest() manifestcap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	if s.base.manifest == nil {
		return nil
	}
	return s.base.manifest.ForPlugin(s.pluginID)
}

// Notifications returns the delegated notification-domain ordinary capability service.
func (s *scopedDirectory) Notifications() capabilitynotifycap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	notifications := s.base.Notifications()
	if scoped, ok := notifications.(notify.CapabilityService); ok {
		return scoped.ForPlugin(s.pluginID)
	}
	return notifications
}

// Org returns the delegated organization capability service.
func (s *scopedDirectory) Org() capabilityorgcap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Org()
}

// Plugins returns the delegated plugin-governance ordinary capability service.
func (s *scopedDirectory) Plugins() capabilityplugincap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.plugins.ForPlugin(s.pluginID)
}

// Route returns the delegated dynamic-route metadata adapter.
func (s *scopedDirectory) Route() routecap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Route()
}

// Sessions returns the delegated online-session domain ordinary capability service.
func (s *scopedDirectory) Sessions() capabilitysessioncap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Sessions()
}

// Storage returns the plugin-scoped host object-storage adapter.
func (s *scopedDirectory) Storage() storagecap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return newStorageAdapter(s.base.storageRuntime, s.base.storageProvider, s.base.bizCtx, s.pluginID)
}

// Tenant returns the delegated tenant capability service.
func (s *scopedDirectory) Tenant() capabilitytenantcap.Service {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Tenant()
}
