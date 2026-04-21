// Package dept exposes the host department controller constructor to source
// plugins without granting direct access to internal packages.
package dept

import (
	deptapi "lina-core/api/dept"
	internaldept "lina-core/internal/controller/dept"
)

// NewV1 creates and returns the host department controller for source-plugin route binding.
func NewV1() deptapi.IDeptV1 {
	return internaldept.NewV1()
}
