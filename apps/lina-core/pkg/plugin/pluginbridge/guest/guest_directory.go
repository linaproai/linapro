// This file implements the guest-side capability directory delegation. The root
// guest.go file owns the public component contract; this file keeps client
// selection details out of the main file.

package guest

import "lina-core/pkg/plugin/capability/recordstore"

// directory implements the guest-side capability directory.
type directory struct{}

// defaultDirectory is the process-default guest-side capability directory.
var defaultDirectory Services = directory{}

// Runtime returns the runtime host service guest client.
func (directory) Runtime() RuntimeHostService { return Runtime() }

// Storage returns the storage host service guest client.
func (directory) Storage() StorageHostService { return Storage() }

// Network returns the outbound network host service guest client.
func (directory) Network() NetworkHostService { return Network() }

// RecordStore returns the governed record store facade for the current dynamic plugin.
func (directory) RecordStore() *recordstore.DB {
	return recordstore.OpenWithHostServiceInvoker(invokeGuestHostService)
}

// Cache returns the cache host service guest client.
func (directory) Cache() CacheHostService { return Cache() }

// Lock returns the distributed lock host service guest client.
func (directory) Lock() LockHostService { return Lock() }

// Plugins returns the plugin-domain guest capability namespace.
func (directory) Plugins() PluginService { return pluginDirectory{} }

// Notify returns the notify host service guest client.
func (directory) Notify() NotifyHostService { return Notify() }

// Cron returns the cron declaration host service guest client.
func (directory) Cron() CronHostService { return Cron() }

// HostConfig returns the host config guest client.
func (directory) HostConfig() HostConfigHostService { return HostConfig() }

// Manifest returns the plugin manifest-resource guest client.
func (directory) Manifest() ManifestHostService { return Manifest() }

// Org returns the organization capability guest client.
func (directory) Org() OrgService { return orgService{} }

// Tenant returns the tenant capability guest client.
func (directory) Tenant() TenantService { return tenantService{} }

// AI returns the guest-side AI capability namespace.
func (directory) AI() AIService { return aiService{} }

// pluginDirectory implements the guest-side plugin-domain namespace.
type pluginDirectory struct{}

// Config returns the plugin-scoped config host service guest client.
func (pluginDirectory) Config() ConfigHostService { return pluginConfig() }
