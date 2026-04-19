// This file centralizes the host SQL directory conventions used by the
// initialization and mock-data commands.

package cmd

import "github.com/gogf/gf/v2/os/gfile"

// hostManifestSqlDir is the canonical directory that stores host SQL delivery files.
const hostManifestSqlDir = "manifest/sql"

// hostInitSqlDir returns the conventional host SQL directory.
func hostInitSqlDir() string {
	return hostManifestSqlDir
}

// hostMockSqlDir returns the conventional host mock-data SQL directory.
func hostMockSqlDir() string {
	return gfile.Join(hostManifestSqlDir, "mock-data")
}
