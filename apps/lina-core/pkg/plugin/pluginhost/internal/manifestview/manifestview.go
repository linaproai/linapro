// Package manifestview implements immutable source-plugin manifest snapshots
// returned through the public pluginhost upgrade callback contract.
package manifestview

import "lina-core/pkg/plugin/capability/capmodel"

// Snapshot is the host-owned immutable manifest snapshot view.
type Snapshot struct {
	value capmodel.ManifestSnapshot
}

// NewSnapshot creates one immutable manifest snapshot view.
func NewSnapshot(value *capmodel.ManifestSnapshot) *Snapshot {
	if value == nil {
		return nil
	}
	return &Snapshot{
		value: *value,
	}
}

// ID returns the plugin identifier recorded in the manifest snapshot.
func (s *Snapshot) ID() string {
	if s == nil {
		return ""
	}
	return s.value.ID
}

// Name returns the plugin display name recorded in the manifest snapshot.
func (s *Snapshot) Name() string {
	if s == nil {
		return ""
	}
	return s.value.Name
}

// Version returns the plugin version recorded in the manifest snapshot.
func (s *Snapshot) Version() string {
	if s == nil {
		return ""
	}
	return s.value.Version
}

// Type returns the plugin type recorded in the manifest snapshot.
func (s *Snapshot) Type() string {
	if s == nil {
		return ""
	}
	return s.value.Type
}

// Values returns a copy of the typed manifest snapshot primitive.
func (s *Snapshot) Values() *capmodel.ManifestSnapshot {
	if s == nil {
		return nil
	}
	value := s.value
	return &value
}
