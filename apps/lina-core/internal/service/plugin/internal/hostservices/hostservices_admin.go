// This file exposes the source-plugin AdminServices directory. Concrete domain
// adapters are wired explicitly by the host owner; nil entries mean the domain
// has no management commands published in the current runtime.

package hostservices

import (
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/configcap"
	"lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/infracap"
	"lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/usercap"
)

// adminDirectory stores typed management surfaces published to source plugins.
type adminDirectory struct {
	users    usercap.AdminService
	auth     authcap.AdminService
	dict     dictcap.AdminService
	files    filecap.AdminService
	sessions sessioncap.AdminService
	config   configcap.AdminService
	notify   notifycap.AdminService
	plugins  plugincap.AdminService
	jobs     jobcap.AdminService
	infra    infracap.AdminService
}

var _ capability.AdminServices = (*adminDirectory)(nil)

// Admin returns the source-plugin management directory.
func (s *directory) Admin() capability.AdminServices {
	if s == nil {
		return nil
	}
	return s.admin
}

// Admin returns the delegated source-plugin management directory.
func (s *scopedDirectory) Admin() capability.AdminServices {
	if s == nil || s.base == nil {
		return nil
	}
	return s.base.Admin()
}

// Users returns user-domain management commands.
func (d *adminDirectory) Users() usercap.AdminService {
	if d == nil {
		return nil
	}
	return d.users
}

// Auth returns authentication and authorization management commands.
func (d *adminDirectory) Auth() authcap.AdminService {
	if d == nil {
		return nil
	}
	return d.auth
}

// Dict returns dictionary-domain management commands.
func (d *adminDirectory) Dict() dictcap.AdminService {
	if d == nil {
		return nil
	}
	return d.dict
}

// Files returns file-domain management commands.
func (d *adminDirectory) Files() filecap.AdminService {
	if d == nil {
		return nil
	}
	return d.files
}

// Sessions returns online-session management commands.
func (d *adminDirectory) Sessions() sessioncap.AdminService {
	if d == nil {
		return nil
	}
	return d.sessions
}

// Config returns runtime configuration management commands.
func (d *adminDirectory) Config() configcap.AdminService {
	if d == nil {
		return nil
	}
	return d.config
}

// Notifications returns notification management commands.
func (d *adminDirectory) Notifications() notifycap.AdminService {
	if d == nil {
		return nil
	}
	return d.notify
}

// Plugins returns plugin-governance management commands.
func (d *adminDirectory) Plugins() plugincap.AdminService {
	if d == nil {
		return nil
	}
	return d.plugins
}

// Jobs returns scheduled-job management commands.
func (d *adminDirectory) Jobs() jobcap.AdminService {
	if d == nil {
		return nil
	}
	return d.jobs
}

// Infra returns infrastructure management commands.
func (d *adminDirectory) Infra() infracap.AdminService {
	if d == nil {
		return nil
	}
	return d.infra
}
