// Package dbdriver registers LinaPro's supported GoFrame SQL drivers and
// provides the shared base driver factory used by host-side wrappers.
package dbdriver

import (
	"strings"

	pgsqlDriver "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/database/gdb"
)

// Supported GoFrame driver type names.
const (
	// TypePostgreSQL is the GoFrame driver type for PostgreSQL connections.
	TypePostgreSQL = "pgsql"
)

// supportedTypes lists the GoFrame driver types registered by this package.
var supportedTypes = []string{TypePostgreSQL}

// SupportedTypes returns a copy of the supported GoFrame driver type names.
func SupportedTypes() []string {
	types := make([]string, len(supportedTypes))
	copy(types, supportedTypes)
	return types
}

// NormalizeType returns one canonical GoFrame driver type name for matching.
func NormalizeType(driverType string) string {
	return strings.ToLower(strings.TrimSpace(driverType))
}

// IsSupported reports whether driverType is registered by LinaPro.
func IsSupported(driverType string) bool {
	switch NormalizeType(driverType) {
	case TypePostgreSQL:
		return true
	default:
		return false
	}
}

// New creates one base GoFrame SQL driver for a supported driver type.
func New(driverType string) (gdb.Driver, bool) {
	switch NormalizeType(driverType) {
	case TypePostgreSQL:
		return pgsqlDriver.New(), true
	default:
		return nil, false
	}
}
