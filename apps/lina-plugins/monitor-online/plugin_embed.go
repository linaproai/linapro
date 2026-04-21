// This file embeds the plugin manifest, frontend pages, and manifest resources.

package monitoronlineplugin

import "embed"

// EmbeddedFiles contains the plugin manifest, frontend pages, and manifest resources.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
