// Package config exposes business-neutral read-only host configuration access
// to source plugins.
package config

import "lina-core/pkg/pluginservice/contract"

// serviceAdapter bridges GoFrame configuration access into the published plugin contract.
type serviceAdapter struct{}

// New creates and returns the published config service adapter.
func New() contract.ConfigService {
	return &serviceAdapter{}
}
