// This file defines lightweight snapshot models used by plugin governance
// persistence.

package store

import (
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// ManifestSnapshot stores the review-friendly manifest snapshot persisted in sys_plugin_release.
type ManifestSnapshot struct {
	ID                        string                      `yaml:"id"`
	Name                      string                      `yaml:"name"`
	Version                   string                      `yaml:"version"`
	Type                      string                      `yaml:"type"`
	ScopeNature               string                      `yaml:"scopeNature,omitempty"`
	SupportsMultiTenant       bool                        `yaml:"supportsMultiTenant,omitempty"`
	DefaultInstallMode        string                      `yaml:"defaultInstallMode,omitempty"`
	Description               string                      `yaml:"description,omitempty"`
	Author                    string                      `yaml:"author,omitempty"`
	Homepage                  string                      `yaml:"homepage,omitempty"`
	License                   string                      `yaml:"license,omitempty"`
	Dependencies              *plugintypes.DependencySpec `yaml:"dependencies,omitempty"`
	RuntimeKind               string                      `yaml:"runtimeKind,omitempty"`
	RuntimeABIVersion         string                      `yaml:"runtimeAbiVersion,omitempty"`
	ManifestDeclared          bool                        `yaml:"manifestDeclared"`
	InstallSQLCount           int                         `yaml:"installSqlCount,omitempty"`
	UninstallSQLCount         int                         `yaml:"uninstallSqlCount,omitempty"`
	MockSQLCount              int                         `yaml:"mockSqlCount,omitempty"`
	FrontendPageCount         int                         `yaml:"frontendPageCount,omitempty"`
	FrontendSlotCount         int                         `yaml:"frontendSlotCount,omitempty"`
	MenuCount                 int                         `yaml:"menuCount,omitempty"`
	BackendHookCount          int                         `yaml:"backendHookCount,omitempty"`
	LifecycleHandlerCount     int                         `yaml:"lifecycleHandlerCount,omitempty"`
	ResourceSpecCount         int                         `yaml:"resourceSpecCount,omitempty"`
	RouteCount                int                         `yaml:"routeCount,omitempty"`
	RouteExecutionEnabled     bool                        `yaml:"routeExecutionEnabled,omitempty"`
	RouteRequestCodec         string                      `yaml:"routeRequestCodec,omitempty"`
	RouteResponseCodec        string                      `yaml:"routeResponseCodec,omitempty"`
	Routes                    []*protocol.RouteContract   `yaml:"routes,omitempty"`
	RuntimeFrontendAssetCount int                         `yaml:"runtimeFrontendAssetCount,omitempty"`
	RuntimeSQLAssetCount      int                         `yaml:"runtimeSqlAssetCount,omitempty"`
	PublicAssets              []*catalog.PublicAssetSpec  `yaml:"public_assets,omitempty"`
	RequestedHostServices     []*protocol.HostServiceSpec `yaml:"requestedHostServices,omitempty"`
	AuthorizedHostServices    []*protocol.HostServiceSpec `yaml:"authorizedHostServices,omitempty"`
	HostServiceAuthRequired   bool                        `yaml:"hostServiceAuthRequired,omitempty"`
	HostServiceAuthConfirmed  bool                        `yaml:"hostServiceAuthConfirmed,omitempty"`
	UninstallPurgeStorageData *bool                       `yaml:"uninstallPurgeStorageData,omitempty"`
}

// PublishedManifestSnapshot converts a persisted manifest snapshot into the
// shared lifecycle callback contract.
func PublishedManifestSnapshot(snapshot *ManifestSnapshot) *protocol.ManifestSnapshotV1 {
	if snapshot == nil {
		return nil
	}
	return &protocol.ManifestSnapshotV1{
		ID:                      snapshot.ID,
		Name:                    snapshot.Name,
		Version:                 snapshot.Version,
		Type:                    snapshot.Type,
		ScopeNature:             snapshot.ScopeNature,
		SupportsMultiTenant:     snapshot.SupportsMultiTenant,
		DefaultInstallMode:      snapshot.DefaultInstallMode,
		Description:             snapshot.Description,
		InstallSQLCount:         snapshot.InstallSQLCount,
		UninstallSQLCount:       snapshot.UninstallSQLCount,
		MockSQLCount:            snapshot.MockSQLCount,
		MenuCount:               snapshot.MenuCount,
		BackendHookCount:        snapshot.BackendHookCount,
		ResourceSpecCount:       snapshot.ResourceSpecCount,
		HostServiceAuthRequired: snapshot.HostServiceAuthRequired,
	}
}
