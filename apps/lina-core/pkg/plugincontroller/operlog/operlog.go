// Package operlog exposes the host operation-log controller constructor to
// source plugins without granting direct access to internal packages.
package operlog

import (
	operlogapi "lina-core/api/operlog"
	internaloperlog "lina-core/internal/controller/operlog"
)

// NewV1 creates and returns the host operation-log controller for source-plugin route binding.
func NewV1() operlogapi.IOperlogV1 {
	return internaloperlog.NewV1()
}
