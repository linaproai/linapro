// Package post exposes the host post controller constructor to source plugins
// without granting direct access to internal packages.
package post

import (
	postapi "lina-core/api/post"
	internalpost "lina-core/internal/controller/post"
)

// NewV1 creates and returns the host post controller for source-plugin route binding.
func NewV1() postapi.IPostV1 {
	return internalpost.NewV1()
}
