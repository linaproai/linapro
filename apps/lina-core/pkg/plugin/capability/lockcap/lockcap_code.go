// This file defines structured plugin lock capability error codes.

package lockcap

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeLockPluginIDRequired reports a lock call before plugin scope binding.
	CodeLockPluginIDRequired = bizerr.MustDefine(
		"PLUGIN_LOCK_PLUGIN_ID_REQUIRED",
		"Plugin lock service requires a plugin ID",
		gcode.CodeInvalidParameter,
	)
	// CodeLockNameRequired reports an empty logical lock name.
	CodeLockNameRequired = bizerr.MustDefine(
		"PLUGIN_LOCK_NAME_REQUIRED",
		"Plugin lock name cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeLockNameTooLong reports a logical or scoped lock name beyond the host limit.
	CodeLockNameTooLong = bizerr.MustDefine(
		"PLUGIN_LOCK_NAME_TOO_LONG",
		"Plugin lock name exceeds the limit of {maxBytes} bytes",
		gcode.CodeInvalidParameter,
	)
	// CodeLockLeaseTooShort reports a lease below MinLease.
	CodeLockLeaseTooShort = bizerr.MustDefine(
		"PLUGIN_LOCK_LEASE_TOO_SHORT",
		"Plugin lock lease cannot be shorter than {minLease}",
		gcode.CodeInvalidParameter,
	)
	// CodeLockLeaseTooLong reports a lease above MaxLease.
	CodeLockLeaseTooLong = bizerr.MustDefine(
		"PLUGIN_LOCK_LEASE_TOO_LONG",
		"Plugin lock lease cannot be longer than {maxLease}",
		gcode.CodeInvalidParameter,
	)
	// CodeLockTicketRequired reports a missing lock ticket for renew or release.
	CodeLockTicketRequired = bizerr.MustDefine(
		"PLUGIN_LOCK_TICKET_REQUIRED",
		"Plugin lock ticket cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeLockServiceUnavailable reports that no shared lock backend is available.
	CodeLockServiceUnavailable = bizerr.MustDefine(
		"PLUGIN_LOCK_SERVICE_UNAVAILABLE",
		"Plugin lock service is unavailable",
		gcode.CodeInternalError,
	)
)
