// This file defines resolver-config business error codes and runtime i18n metadata.

package resolverconfig

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeResolverConfigInvalid reports invalid resolver configuration input.
	CodeResolverConfigInvalid = bizerr.MustDefine(
		"MULTI_TENANT_RESOLVER_CONFIG_INVALID",
		"Tenant resolver configuration is invalid",
		gcode.CodeInvalidParameter,
	)
)
