// Package loginlog exposes the host login-log controller constructor to source
// plugins without granting direct access to internal packages.
package loginlog

import (
	loginlogapi "lina-core/api/loginlog"
	internalloginlog "lina-core/internal/controller/loginlog"
)

// NewV1 creates and returns the host login-log controller for source-plugin route binding.
func NewV1() loginlogapi.ILoginlogV1 {
	return internalloginlog.NewV1()
}
