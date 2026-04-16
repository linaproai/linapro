package main

import "embed"

// EmbeddedFiles contains the plugin manifest, hosted frontend assets, and
// convention-based SQL resources used to generate runtime snapshot sections.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
