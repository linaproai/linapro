// This file exposes shared WASM custom-section readers for dynamic plugin
// artifact discovery.

package artifact

import (
	"strings"

	"lina-core/pkg/pluginbridge/artifact/internal/wasmsection"
)

// ReadCustomSection returns one named WASM custom section payload.
func ReadCustomSection(content []byte, name string) ([]byte, bool, error) {
	sections, err := ListCustomSections(content)
	if err != nil {
		return nil, false, err
	}
	payload, ok := sections[strings.TrimSpace(name)]
	if !ok {
		return nil, false, nil
	}
	return payload, true, nil
}

// ListCustomSections extracts all WASM custom section payloads by section name.
func ListCustomSections(content []byte) (map[string][]byte, error) {
	return wasmsection.ListCustomSections(content)
}
