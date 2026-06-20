// This file defines source-plugin lifecycle and hook input wrappers that
// isolate plugin callbacks from host-internal models.

package pluginhost

import "lina-core/pkg/plugin/pluginhost/internal/valuecopy"

// HookPayload exposes one published host hook payload.
type HookPayload interface {
	// ExtensionPoint returns the published extension point of the current callback.
	ExtensionPoint() ExtensionPoint
	// Value returns one published payload field by key.
	Value(key string) interface{}
	// Values returns a copy of all published payload fields.
	Values() map[string]interface{}
	// Services returns the host-published runtime services for hook handlers.
	Services() Services
}

// hookPayload is the host-owned implementation of the published HookPayload view.
type hookPayload struct {
	point    ExtensionPoint
	values   map[string]interface{}
	services Services
}

// SourcePluginUninstallInput exposes one host-confirmed uninstall policy snapshot to a source plugin.
type SourcePluginUninstallInput interface {
	// PluginID returns the source-plugin identifier being uninstalled.
	PluginID() string
	// PurgeStorageData reports whether the host expects the plugin to clear its
	// own business data and stored files during uninstall.
	PurgeStorageData() bool
	// Services returns the plugin-scoped host service directory available during
	// uninstall cleanup. It can be nil for tests that construct callbacks
	// outside the host lifecycle runner.
	Services() Services
}

// SourcePluginLifecycleInput exposes one generic plugin lifecycle operation to
// source-plugin precondition callbacks.
type SourcePluginLifecycleInput interface {
	// PluginID returns the source-plugin identifier targeted by the lifecycle action.
	PluginID() string
	// Operation returns the stable lifecycle operation key.
	Operation() string
	// StartupAutoEnable reports whether this lifecycle action was initiated by
	// the host startup plugin.autoEnable bootstrap for the target plugin.
	StartupAutoEnable() bool
	// PurgeStorageData reports whether an uninstall lifecycle action should
	// clear plugin-owned storage and business data. It returns false for
	// non-uninstall lifecycle operations.
	PurgeStorageData() bool
}

// SourcePluginTenantLifecycleInput exposes one tenant lifecycle operation to
// source-plugin precondition callbacks.
type SourcePluginTenantLifecycleInput interface {
	// Operation returns the stable lifecycle operation key.
	Operation() string
	// TenantID returns the target tenant identifier.
	TenantID() int
}

// SourcePluginInstallModeChangeInput exposes one plugin install-mode transition
// to source-plugin precondition callbacks.
type SourcePluginInstallModeChangeInput interface {
	// PluginID returns the source-plugin identifier targeted by the transition.
	PluginID() string
	// Operation returns the stable lifecycle operation key.
	Operation() string
	// FromMode returns the current install mode.
	FromMode() string
	// ToMode returns the requested install mode.
	ToMode() string
}

// SourcePluginUpgradeInput exposes one host-confirmed source-plugin runtime
// upgrade request to BeforeUpgrade, Upgrade, and AfterUpgrade callbacks.
type SourcePluginUpgradeInput interface {
	// PluginID returns the source-plugin identifier being upgraded.
	PluginID() string
	// FromVersion returns the effective version before the upgrade request.
	FromVersion() string
	// ToVersion returns the target version discovered from the new manifest.
	ToVersion() string
	// FromManifest returns the effective manifest snapshot before upgrade.
	FromManifest() ManifestSnapshot
	// ToManifest returns the target manifest snapshot after upgrade.
	ToManifest() ManifestSnapshot
}

// sourcePluginUninstallInput is the host-owned implementation of the published
// uninstall policy snapshot passed to source plugins.
type sourcePluginUninstallInput struct {
	pluginID         string
	purgeStorageData bool
	services         Services
}

// sourcePluginLifecycleInput is the host-owned implementation passed to generic
// source-plugin lifecycle precondition callbacks.
type sourcePluginLifecycleInput struct {
	pluginID          string
	operation         string
	startupAutoEnable bool
	purgeStorageData  bool
}

// sourcePluginTenantLifecycleInput is the host-owned implementation passed to
// tenant lifecycle precondition callbacks.
type sourcePluginTenantLifecycleInput struct {
	operation string
	tenantID  int
}

// sourcePluginInstallModeChangeInput is the host-owned implementation passed to
// install-mode change precondition callbacks.
type sourcePluginInstallModeChangeInput struct {
	pluginID  string
	operation string
	fromMode  string
	toMode    string
}

// sourcePluginUpgradeInput is the host-owned implementation passed to source
// plugin upgrade lifecycle callbacks.
type sourcePluginUpgradeInput struct {
	pluginID     string
	fromVersion  string
	toVersion    string
	fromManifest ManifestSnapshot
	toManifest   ManifestSnapshot
}

// NewHookPayload creates one published hook payload wrapper for plugins.
func NewHookPayload(point ExtensionPoint, values map[string]interface{}) HookPayload {
	return NewHookPayloadWithServices(point, values, nil)
}

// NewHookPayloadWithServices creates one published hook payload wrapper with
// host-published runtime services available to source plugins.
func NewHookPayloadWithServices(
	point ExtensionPoint,
	values map[string]interface{},
	services Services,
) HookPayload {
	return &hookPayload{
		point:    point,
		values:   valuecopy.Map(values),
		services: services,
	}
}

// NewSourcePluginUninstallInput creates one published source-plugin uninstall input wrapper.
func NewSourcePluginUninstallInput(
	pluginID string,
	purgeStorageData bool,
) SourcePluginUninstallInput {
	return NewSourcePluginUninstallInputWithServices(pluginID, purgeStorageData, nil)
}

// NewSourcePluginUninstallInputWithServices creates one source-plugin
// uninstall input wrapper with a plugin-scoped host service directory.
func NewSourcePluginUninstallInputWithServices(
	pluginID string,
	purgeStorageData bool,
	services Services,
) SourcePluginUninstallInput {
	return &sourcePluginUninstallInput{
		pluginID:         pluginID,
		purgeStorageData: purgeStorageData,
		services:         services,
	}
}

// NewSourcePluginLifecycleInput creates one published generic lifecycle input wrapper.
func NewSourcePluginLifecycleInput(pluginID string, operation string) SourcePluginLifecycleInput {
	return &sourcePluginLifecycleInput{
		pluginID:  pluginID,
		operation: operation,
	}
}

// NewSourcePluginLifecycleInputWithUninstallPolicy creates one lifecycle input
// wrapper with the host-confirmed uninstall cleanup policy attached.
func NewSourcePluginLifecycleInputWithUninstallPolicy(
	pluginID string,
	operation string,
	purgeStorageData bool,
) SourcePluginLifecycleInput {
	return NewSourcePluginLifecycleInputWithPolicy(pluginID, operation, SourcePluginLifecyclePolicy{
		PurgeStorageData: purgeStorageData,
	})
}

// SourcePluginLifecyclePolicy carries host-owned lifecycle metadata that source
// plugins can inspect without depending on internal plugin facade options.
type SourcePluginLifecyclePolicy struct {
	// StartupAutoEnable reports whether the action was initiated by startup
	// plugin.autoEnable for the target plugin.
	StartupAutoEnable bool
	// PurgeStorageData reports whether uninstall should remove plugin-owned data.
	PurgeStorageData bool
}

// NewSourcePluginLifecycleInputWithPolicy creates one lifecycle input wrapper
// with host-owned lifecycle metadata attached.
func NewSourcePluginLifecycleInputWithPolicy(
	pluginID string,
	operation string,
	policy SourcePluginLifecyclePolicy,
) SourcePluginLifecycleInput {
	return &sourcePluginLifecycleInput{
		pluginID:          pluginID,
		operation:         operation,
		startupAutoEnable: policy.StartupAutoEnable,
		purgeStorageData:  policy.PurgeStorageData,
	}
}

// NewSourcePluginTenantLifecycleInput creates one published tenant lifecycle input wrapper.
func NewSourcePluginTenantLifecycleInput(operation string, tenantID int) SourcePluginTenantLifecycleInput {
	return &sourcePluginTenantLifecycleInput{
		operation: operation,
		tenantID:  tenantID,
	}
}

// NewSourcePluginInstallModeChangeInput creates one published install-mode change input wrapper.
func NewSourcePluginInstallModeChangeInput(
	pluginID string,
	operation string,
	fromMode string,
	toMode string,
) SourcePluginInstallModeChangeInput {
	return &sourcePluginInstallModeChangeInput{
		pluginID:  pluginID,
		operation: operation,
		fromMode:  fromMode,
		toMode:    toMode,
	}
}

// NewSourcePluginUpgradeInput creates one published source-plugin upgrade input wrapper.
func NewSourcePluginUpgradeInput(
	pluginID string,
	fromVersion string,
	toVersion string,
	fromManifest ManifestSnapshot,
	toManifest ManifestSnapshot,
) SourcePluginUpgradeInput {
	return &sourcePluginUpgradeInput{
		pluginID:     pluginID,
		fromVersion:  fromVersion,
		toVersion:    toVersion,
		fromManifest: fromManifest,
		toManifest:   toManifest,
	}
}

// ExtensionPoint returns the published extension point of the current hook payload.
func (p *hookPayload) ExtensionPoint() ExtensionPoint {
	if p == nil {
		return ""
	}
	return p.point
}

// Value returns one published payload field by key.
func (p *hookPayload) Value(key string) interface{} {
	if p == nil {
		return nil
	}
	return p.values[key]
}

// Values returns a shallow copy of all published payload fields.
func (p *hookPayload) Values() map[string]interface{} {
	if p == nil {
		return map[string]interface{}{}
	}
	return valuecopy.Map(p.values)
}

// Services returns the host-published runtime services for hook handlers.
func (p *hookPayload) Services() Services {
	if p == nil {
		return nil
	}
	return p.services
}

// PluginID returns the source-plugin identifier being uninstalled.
func (i *sourcePluginUninstallInput) PluginID() string {
	if i == nil {
		return ""
	}
	return i.pluginID
}

// PurgeStorageData reports whether the host expects business data cleanup.
func (i *sourcePluginUninstallInput) PurgeStorageData() bool {
	if i == nil {
		return false
	}
	return i.purgeStorageData
}

// Services returns the plugin-scoped host services available to uninstall cleanup.
func (i *sourcePluginUninstallInput) Services() Services {
	if i == nil {
		return nil
	}
	return i.services
}

// PluginID returns the source-plugin identifier for a generic lifecycle input.
func (i *sourcePluginLifecycleInput) PluginID() string {
	if i == nil {
		return ""
	}
	return i.pluginID
}

// Operation returns the lifecycle operation key.
func (i *sourcePluginLifecycleInput) Operation() string {
	if i == nil {
		return ""
	}
	return i.operation
}

// StartupAutoEnable reports whether the host started this action from config.
func (i *sourcePluginLifecycleInput) StartupAutoEnable() bool {
	if i == nil {
		return false
	}
	return i.startupAutoEnable
}

// PurgeStorageData reports whether an uninstall lifecycle should clear data.
func (i *sourcePluginLifecycleInput) PurgeStorageData() bool {
	if i == nil {
		return false
	}
	return i.purgeStorageData
}

// Operation returns the tenant lifecycle operation key.
func (i *sourcePluginTenantLifecycleInput) Operation() string {
	if i == nil {
		return ""
	}
	return i.operation
}

// TenantID returns the target tenant identifier.
func (i *sourcePluginTenantLifecycleInput) TenantID() int {
	if i == nil {
		return 0
	}
	return i.tenantID
}

// PluginID returns the source-plugin identifier for an install-mode transition.
func (i *sourcePluginInstallModeChangeInput) PluginID() string {
	if i == nil {
		return ""
	}
	return i.pluginID
}

// Operation returns the install-mode lifecycle operation key.
func (i *sourcePluginInstallModeChangeInput) Operation() string {
	if i == nil {
		return ""
	}
	return i.operation
}

// FromMode returns the current install mode.
func (i *sourcePluginInstallModeChangeInput) FromMode() string {
	if i == nil {
		return ""
	}
	return i.fromMode
}

// ToMode returns the requested install mode.
func (i *sourcePluginInstallModeChangeInput) ToMode() string {
	if i == nil {
		return ""
	}
	return i.toMode
}

// PluginID returns the source-plugin identifier being upgraded.
func (i *sourcePluginUpgradeInput) PluginID() string {
	if i == nil {
		return ""
	}
	return i.pluginID
}

// FromVersion returns the effective source-plugin version before upgrade.
func (i *sourcePluginUpgradeInput) FromVersion() string {
	if i == nil {
		return ""
	}
	return i.fromVersion
}

// ToVersion returns the target source-plugin version after upgrade.
func (i *sourcePluginUpgradeInput) ToVersion() string {
	if i == nil {
		return ""
	}
	return i.toVersion
}

// FromManifest returns the effective manifest snapshot before upgrade.
func (i *sourcePluginUpgradeInput) FromManifest() ManifestSnapshot {
	if i == nil {
		return nil
	}
	return i.fromManifest
}

// ToManifest returns the target manifest snapshot after upgrade.
func (i *sourcePluginUpgradeInput) ToManifest() ManifestSnapshot {
	if i == nil {
		return nil
	}
	return i.toManifest
}
