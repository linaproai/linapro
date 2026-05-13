// This file embeds the media plugin manifest, frontend pages, and manifest resources.

package media

import "embed"

// EmbeddedFiles contains the plugin manifest, frontend pages, and manifest resources.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
