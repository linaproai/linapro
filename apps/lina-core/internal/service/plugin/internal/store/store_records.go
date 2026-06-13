// This file defines stable store-owned projections for plugin governance rows.

package store

import "time"

// PluginRecord is the store-owned projection of one sys_plugin row.
type PluginRecord struct {
	Id                      int
	PluginId                string
	Name                    string
	Version                 string
	Type                    string
	Installed               int
	Status                  int
	DesiredState            string
	CurrentState            string
	Generation              int64
	ReleaseId               int
	ManifestPath            string
	Checksum                string
	InstalledAt             *time.Time
	EnabledAt               *time.Time
	DisabledAt              *time.Time
	Remark                  string
	ScopeNature             string
	InstallMode             string
	AutoEnableForNewTenants bool
	CreatedAt               *time.Time
	UpdatedAt               *time.Time
}

// ReleaseRecord is the store-owned projection of one sys_plugin_release row.
type ReleaseRecord struct {
	Id               int
	PluginId         string
	ReleaseVersion   string
	Type             string
	RuntimeKind      string
	SchemaVersion    string
	MinHostVersion   string
	MaxHostVersion   string
	Status           string
	ManifestPath     string
	PackagePath      string
	Checksum         string
	ManifestSnapshot string
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
}

// MigrationRecord is the store-owned projection of one sys_plugin_migration row.
type MigrationRecord struct {
	Id           int
	PluginId     string
	ReleaseId    int
	Phase        string
	MigrationKey string
	Status       string
	ErrorMessage string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

// NodeStateRecord is the store-owned projection of one sys_plugin_node_state row.
type NodeStateRecord struct {
	Id              int
	PluginId        string
	ReleaseId       int
	NodeKey         string
	DesiredState    string
	CurrentState    string
	Generation      int64
	LastHeartbeatAt *time.Time
	ErrorMessage    string
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
}
