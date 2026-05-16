// This file defines manifest snapshot wrappers published to source-plugin
// upgrade callbacks.

package pluginhost

import bridgecontract "lina-core/pkg/pluginbridge/contract"

// ManifestSnapshot exposes the review-oriented manifest snapshot fields needed
// by source-plugin upgrade callbacks without leaking host catalog internals.
type ManifestSnapshot interface {
	// ID returns the plugin identifier recorded in the manifest snapshot.
	ID() string
	// Name returns the plugin display name recorded in the manifest snapshot.
	Name() string
	// Version returns the plugin version recorded in the manifest snapshot.
	Version() string
	// Type returns the plugin type recorded in the manifest snapshot.
	Type() string
	// Values returns the typed bridge manifest snapshot contract.
	Values() *bridgecontract.ManifestSnapshotV1
}

// manifestSnapshot is the host-owned immutable view passed to source-plugin
// runtime upgrade callbacks.
type manifestSnapshot struct {
	value bridgecontract.ManifestSnapshotV1
}

// NewManifestSnapshot creates one published manifest snapshot wrapper from the
// shared lifecycle callback contract.
func NewManifestSnapshot(value *bridgecontract.ManifestSnapshotV1) ManifestSnapshot {
	if value == nil {
		return nil
	}
	return &manifestSnapshot{
		value: *value,
	}
}

// ID returns the plugin identifier recorded in the manifest snapshot.
func (s *manifestSnapshot) ID() string {
	if s == nil {
		return ""
	}
	return s.value.ID
}

// Name returns the plugin display name recorded in the manifest snapshot.
func (s *manifestSnapshot) Name() string {
	if s == nil {
		return ""
	}
	return s.value.Name
}

// Version returns the plugin version recorded in the manifest snapshot.
func (s *manifestSnapshot) Version() string {
	if s == nil {
		return ""
	}
	return s.value.Version
}

// Type returns the plugin type recorded in the manifest snapshot.
func (s *manifestSnapshot) Type() string {
	if s == nil {
		return ""
	}
	return s.value.Type
}

// Values returns a copy of the shared typed manifest snapshot contract.
func (s *manifestSnapshot) Values() *bridgecontract.ManifestSnapshotV1 {
	if s == nil {
		return nil
	}
	value := s.value
	return &value
}
