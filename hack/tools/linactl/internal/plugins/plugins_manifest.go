// This file classifies plugin manifests for dynamic Wasm builds and keeps
// manifest parsing reusable across workspace generation and status checks.

package plugins

import (
	"strings"
)

// IsDynamic reports whether a plugin manifest declares dynamic type.
func IsDynamic(manifestPath string) (bool, error) {
	manifest, err := ReadManifest(manifestPath)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(manifest.Type), "dynamic"), nil
}
