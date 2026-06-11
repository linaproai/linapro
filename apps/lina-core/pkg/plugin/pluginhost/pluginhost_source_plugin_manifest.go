// This file defines manifest snapshot wrappers published to source-plugin
// upgrade callbacks.

package pluginhost

import (
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/pluginhost/internal/manifestview"
)

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
	// Values returns the typed manifest snapshot primitive.
	Values() *capmodel.ManifestSnapshot
}

// NewManifestSnapshot creates one published manifest snapshot wrapper from the
// shared lifecycle callback contract.
func NewManifestSnapshot(value *capmodel.ManifestSnapshot) ManifestSnapshot {
	if value == nil {
		return nil
	}
	return manifestview.NewSnapshot(value)
}
