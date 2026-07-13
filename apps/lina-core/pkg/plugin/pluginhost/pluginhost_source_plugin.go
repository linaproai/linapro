// This file defines the host-owned source-plugin registration storage used by
// grouped source-plugin facade implementations.

package pluginhost

import (
	"io/fs"

	"lina-core/pkg/plugin/capability/authcap/extlogin/extidspi"
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// sourcePlugin stores one compile-time source plugin definition behind the
// published grouped Declarations interface.
type sourcePlugin struct {
	// id is the stable plugin id and must match `plugin.yaml`.
	id string
	// assets exposes grouped asset registration helpers.
	assets AssetDeclarations
	// lifecycle exposes grouped lifecycle registration helpers.
	lifecycle LifecycleDeclarations
	// hooks exposes grouped hook registration helpers.
	hooks HookDeclarations
	// http exposes grouped HTTP registration helpers.
	http HTTPDeclarations
	// jobs exposes grouped scheduled-job registration helpers.
	jobs JobDeclarations
	// providers exposes grouped framework provider declaration helpers.
	providers ProviderDeclarations
	// access exposes grouped menu and permission access-control helpers.
	access AccessDeclarations

	embeddedFiles          fs.FS
	tenantProvider         tenantspi.ProviderFactory
	orgProvider            orgspi.ProviderFactory
	capabilities           []capregistry.Descriptor
	externalIdentities     []string
	externalIdentityEngine extidspi.ProviderFactory
	beforeInstall          SourcePluginBeforeLifecycleHandler
	afterInstall           SourcePluginAfterLifecycleHandler
	beforeEnable           SourcePluginBeforeLifecycleHandler
	afterEnable            SourcePluginAfterLifecycleHandler
	beforeUpgrade          SourcePluginBeforeUpgradeHandler
	upgradeHandler         SourcePluginUpgradeHandler
	afterUpgrade           SourcePluginUpgradeHandler
	beforeDisable          SourcePluginBeforeLifecycleHandler
	afterDisable           SourcePluginAfterLifecycleHandler
	beforeUninstall        SourcePluginBeforeLifecycleHandler
	afterUninstall         SourcePluginAfterLifecycleHandler
	globalBeforeInstall    SourcePluginGlobalLifecycleHandler
	globalBeforeEnable     SourcePluginGlobalLifecycleHandler
	globalBeforeDisable    SourcePluginGlobalLifecycleHandler
	globalBeforeUninstall  SourcePluginGlobalLifecycleHandler
	beforeTenantDis        SourcePluginBeforeTenantLifecycleHandler
	afterTenantDis         SourcePluginAfterTenantLifecycleHandler
	beforeTenantDel        SourcePluginBeforeTenantLifecycleHandler
	afterTenantDel         SourcePluginAfterTenantLifecycleHandler
	beforeModeChange       SourcePluginBeforeInstallModeChangeHandler
	afterModeChange        SourcePluginAfterInstallModeChangeHandler
	uninstallHandler       SourcePluginUninstallHandler
	hookHandlers           []*HookHandlerRegistration
	routeRegistrars        []*RouteHandlerRegistration
	jobRegistrars          []*JobHandlerRegistration
	menuFilters            []*MenuFilterHandlerRegistration
	permissionFilters      []*PermissionFilterHandlerRegistration
}

// NewDeclarations creates and returns a new grouped source-plugin declarations facade.
func NewDeclarations(id string) Declarations {
	plugin := &sourcePlugin{
		id:                id,
		capabilities:      make([]capregistry.Descriptor, 0),
		hookHandlers:      make([]*HookHandlerRegistration, 0),
		routeRegistrars:   make([]*RouteHandlerRegistration, 0),
		jobRegistrars:     make([]*JobHandlerRegistration, 0),
		menuFilters:       make([]*MenuFilterHandlerRegistration, 0),
		permissionFilters: make([]*PermissionFilterHandlerRegistration, 0),
	}
	plugin.assets = &sourcePluginAssets{plugin: plugin}
	plugin.lifecycle = &sourcePluginLifecycle{plugin: plugin}
	plugin.hooks = &sourcePluginHooks{plugin: plugin}
	plugin.http = &sourcePluginHTTP{plugin: plugin}
	plugin.jobs = &sourcePluginJobs{plugin: plugin}
	plugin.providers = &sourcePluginProviders{plugin: plugin}
	plugin.access = &sourcePluginAccess{plugin: plugin}
	return plugin
}

// useEmbeddedFiles binds one plugin-owned embedded filesystem to the source plugin.
func (p *sourcePlugin) useEmbeddedFiles(fileSystem fs.FS) {
	if p == nil {
		return
	}
	p.embeddedFiles = fileSystem
}

// GetEmbeddedFiles returns the plugin-owned embedded filesystem when declared.
func (p *sourcePlugin) GetEmbeddedFiles() fs.FS {
	if p == nil {
		return nil
	}
	return p.embeddedFiles
}

// GetHookHandlers returns the registered callback-style hook handlers.
func (p *sourcePlugin) GetHookHandlers() []*HookHandlerRegistration {
	if p == nil {
		return []*HookHandlerRegistration{}
	}
	items := make([]*HookHandlerRegistration, len(p.hookHandlers))
	copy(items, p.hookHandlers)
	return items
}

// GetRouteRegistrars returns the registered route contribution callbacks.
func (p *sourcePlugin) GetRouteRegistrars() []*RouteHandlerRegistration {
	if p == nil {
		return []*RouteHandlerRegistration{}
	}
	items := make([]*RouteHandlerRegistration, len(p.routeRegistrars))
	copy(items, p.routeRegistrars)
	return items
}

// GetJobRegistrars returns the registered scheduled-job contribution callbacks.
func (p *sourcePlugin) GetJobRegistrars() []*JobHandlerRegistration {
	if p == nil {
		return []*JobHandlerRegistration{}
	}
	items := make([]*JobHandlerRegistration, len(p.jobRegistrars))
	copy(items, p.jobRegistrars)
	return items
}

// GetMenuFilters returns the registered menu filter callbacks.
func (p *sourcePlugin) GetMenuFilters() []*MenuFilterHandlerRegistration {
	if p == nil {
		return []*MenuFilterHandlerRegistration{}
	}
	items := make([]*MenuFilterHandlerRegistration, len(p.menuFilters))
	copy(items, p.menuFilters)
	return items
}

// GetPermissionFilters returns the registered permission filter callbacks.
func (p *sourcePlugin) GetPermissionFilters() []*PermissionFilterHandlerRegistration {
	if p == nil {
		return []*PermissionFilterHandlerRegistration{}
	}
	items := make([]*PermissionFilterHandlerRegistration, len(p.permissionFilters))
	copy(items, p.permissionFilters)
	return items
}

// GetTenantProviderFactory returns the declared tenant provider factory.
func (p *sourcePlugin) GetTenantProviderFactory() tenantspi.ProviderFactory {
	if p == nil {
		return nil
	}
	return p.tenantProvider
}

// GetOrgProviderFactory returns the declared organization provider factory.
func (p *sourcePlugin) GetOrgProviderFactory() orgspi.ProviderFactory {
	if p == nil {
		return nil
	}
	return p.orgProvider
}

// GetCapabilityDescriptors returns plugin-owned capability descriptors declared by this source plugin.
func (p *sourcePlugin) GetCapabilityDescriptors() []capregistry.Descriptor {
	if p == nil {
		return []capregistry.Descriptor{}
	}
	items := make([]capregistry.Descriptor, 0, len(p.capabilities))
	for _, descriptor := range p.capabilities {
		items = append(items, cloneCapabilityDescriptor(descriptor))
	}
	return items
}

// GetExternalIdentityProviderFactory returns the declared external-identity
// provider engine factory. It is distinct from GetExternalIdentityProviderIDs:
// the factory supplies the resolve/provision engine (declared by
// linapro-extlogin-core), while the ID list stamps ownership for calling plugins.
func (p *sourcePlugin) GetExternalIdentityProviderFactory() extidspi.ProviderFactory {
	if p == nil {
		return nil
	}
	return p.externalIdentityEngine
}

func cloneCapabilityDescriptor(descriptor capregistry.Descriptor) capregistry.Descriptor {
	descriptor.Methods = append([]capregistry.MethodDescriptor(nil), descriptor.Methods...)
	return descriptor
}

// GetExternalIdentityProviderIDs returns a copy of the external-identity
// provider IDs declared by this source plugin.
func (p *sourcePlugin) GetExternalIdentityProviderIDs() []string {
	if p == nil || len(p.externalIdentities) == 0 {
		return nil
	}
	items := make([]string, len(p.externalIdentities))
	copy(items, p.externalIdentities)
	return items
}

// GetBeforeInstallHandler returns the registered source-plugin pre-install callback.
func (p *sourcePlugin) GetBeforeInstallHandler() SourcePluginBeforeLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.beforeInstall
}

// GetAfterInstallHandler returns the registered source-plugin post-install callback.
func (p *sourcePlugin) GetAfterInstallHandler() SourcePluginAfterLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.afterInstall
}

// GetBeforeEnableHandler returns the registered source-plugin pre-enable callback.
func (p *sourcePlugin) GetBeforeEnableHandler() SourcePluginBeforeLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.beforeEnable
}

// GetAfterEnableHandler returns the registered source-plugin post-enable callback.
func (p *sourcePlugin) GetAfterEnableHandler() SourcePluginAfterLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.afterEnable
}

// GetGlobalBeforeInstallHandler returns the registered global pre-install callback.
func (p *sourcePlugin) GetGlobalBeforeInstallHandler() SourcePluginGlobalLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.globalBeforeInstall
}

// GetGlobalBeforeEnableHandler returns the registered global pre-enable callback.
func (p *sourcePlugin) GetGlobalBeforeEnableHandler() SourcePluginGlobalLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.globalBeforeEnable
}

// GetGlobalBeforeDisableHandler returns the registered global pre-disable callback.
func (p *sourcePlugin) GetGlobalBeforeDisableHandler() SourcePluginGlobalLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.globalBeforeDisable
}

// GetGlobalBeforeUninstallHandler returns the registered global pre-uninstall callback.
func (p *sourcePlugin) GetGlobalBeforeUninstallHandler() SourcePluginGlobalLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.globalBeforeUninstall
}

// GetBeforeUpgradeHandler returns the registered source-plugin pre-upgrade callback.
func (p *sourcePlugin) GetBeforeUpgradeHandler() SourcePluginBeforeUpgradeHandler {
	if p == nil {
		return nil
	}
	return p.beforeUpgrade
}

// GetUpgradeHandler returns the registered source-plugin custom upgrade callback.
func (p *sourcePlugin) GetUpgradeHandler() SourcePluginUpgradeHandler {
	if p == nil {
		return nil
	}
	return p.upgradeHandler
}

// GetAfterUpgradeHandler returns the registered source-plugin post-upgrade callback.
func (p *sourcePlugin) GetAfterUpgradeHandler() SourcePluginUpgradeHandler {
	if p == nil {
		return nil
	}
	return p.afterUpgrade
}

// GetBeforeDisableHandler returns the registered source-plugin pre-disable callback.
func (p *sourcePlugin) GetBeforeDisableHandler() SourcePluginBeforeLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.beforeDisable
}

// GetAfterDisableHandler returns the registered source-plugin post-disable callback.
func (p *sourcePlugin) GetAfterDisableHandler() SourcePluginAfterLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.afterDisable
}

// GetBeforeUninstallHandler returns the registered source-plugin pre-uninstall callback.
func (p *sourcePlugin) GetBeforeUninstallHandler() SourcePluginBeforeLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.beforeUninstall
}

// GetAfterUninstallHandler returns the registered source-plugin post-uninstall callback.
func (p *sourcePlugin) GetAfterUninstallHandler() SourcePluginAfterLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.afterUninstall
}

// GetBeforeTenantDisableHandler returns the registered source-plugin tenant-disable callback.
func (p *sourcePlugin) GetBeforeTenantDisableHandler() SourcePluginBeforeTenantLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.beforeTenantDis
}

// GetAfterTenantDisableHandler returns the registered source-plugin post-tenant-disable callback.
func (p *sourcePlugin) GetAfterTenantDisableHandler() SourcePluginAfterTenantLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.afterTenantDis
}

// GetBeforeTenantDeleteHandler returns the registered source-plugin tenant-delete callback.
func (p *sourcePlugin) GetBeforeTenantDeleteHandler() SourcePluginBeforeTenantLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.beforeTenantDel
}

// GetAfterTenantDeleteHandler returns the registered source-plugin post-tenant-delete callback.
func (p *sourcePlugin) GetAfterTenantDeleteHandler() SourcePluginAfterTenantLifecycleHandler {
	if p == nil {
		return nil
	}
	return p.afterTenantDel
}

// GetBeforeInstallModeChangeHandler returns the registered source-plugin install-mode callback.
func (p *sourcePlugin) GetBeforeInstallModeChangeHandler() SourcePluginBeforeInstallModeChangeHandler {
	if p == nil {
		return nil
	}
	return p.beforeModeChange
}

// GetAfterInstallModeChangeHandler returns the registered source-plugin post-install-mode callback.
func (p *sourcePlugin) GetAfterInstallModeChangeHandler() SourcePluginAfterInstallModeChangeHandler {
	if p == nil {
		return nil
	}
	return p.afterModeChange
}

// GetUninstallHandler returns the registered source-plugin uninstall cleanup callback.
func (p *sourcePlugin) GetUninstallHandler() SourcePluginUninstallHandler {
	if p == nil {
		return nil
	}
	return p.uninstallHandler
}
