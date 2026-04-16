package packed

import "embed"

// Files stores embedded frontend static assets and prepared manifest assets.
//
//go:embed all:public all:manifest
var Files embed.FS
