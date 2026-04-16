package plugindemosource

import "embed"

// EmbeddedFiles contains the plugin manifest, convention-based SQL assets, and frontend source resources.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
