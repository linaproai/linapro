// This file defines tenant business error codes and runtime i18n metadata.

package tenant

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeTenantNotFound reports that the requested tenant does not exist.
	CodeTenantNotFound = bizerr.MustDefine(
		"MULTI_TENANT_TENANT_NOT_FOUND",
		"Tenant does not exist",
		gcode.CodeNotFound,
	)
	// CodeTenantCodeExists reports that a tenant code is already used.
	CodeTenantCodeExists = bizerr.MustDefine(
		"MULTI_TENANT_TENANT_CODE_EXISTS",
		"Tenant code already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeTenantCodeInvalid reports that a tenant code does not match the public code contract.
	CodeTenantCodeInvalid = bizerr.MustDefine(
		"MULTI_TENANT_TENANT_CODE_INVALID",
		"Tenant code must use lowercase letters, numbers, and hyphens with length 2 to 32",
		gcode.CodeInvalidParameter,
	)
	// CodeTenantCodeReserved reports that a deleted tenant still reserves its code tombstone.
	CodeTenantCodeReserved = bizerr.MustDefine(
		"MULTI_TENANT_TENANT_CODE_RESERVED",
		"Tenant code is reserved by a deleted tenant",
		gcode.CodeInvalidParameter,
	)
	// CodeTenantInvalidStatus reports an unsupported tenant lifecycle status.
	CodeTenantInvalidStatus = bizerr.MustDefine(
		"MULTI_TENANT_INVALID_TENANT_STATUS",
		"Tenant status is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeTenantStatusTransitionInvalid reports a forbidden lifecycle transition.
	CodeTenantStatusTransitionInvalid = bizerr.MustDefine(
		"MULTI_TENANT_TENANT_STATUS_TRANSITION_INVALID",
		"Tenant status transition is invalid",
		gcode.CodeInvalidOperation,
	)
	// CodeTenantDeleteGuardVetoed reports that a plugin lifecycle guard blocked tenant deletion.
	CodeTenantDeleteGuardVetoed = bizerr.MustDefine(
		"MULTI_TENANT_DELETE_GUARD_VETOED",
		"Tenant deletion was blocked by a lifecycle guard",
		gcode.CodeInvalidOperation,
	)
)
