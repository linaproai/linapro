// This file defines plugin host-lock business error codes and their i18n
// metadata.

package hostlock

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeHostLockPluginIDRequired reports that the caller plugin ID is required.
	CodeHostLockPluginIDRequired = bizerr.MustDefine(
		"HOST_LOCK_PLUGIN_ID_REQUIRED",
		"Plugin ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeHostLockResourceRequired reports that the logical lock resource name is required.
	CodeHostLockResourceRequired = bizerr.MustDefine(
		"HOST_LOCK_RESOURCE_REQUIRED",
		"Logical lock resource cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeHostLockNameTooLong reports that the actual lock name exceeds the storage limit.
	CodeHostLockNameTooLong = bizerr.MustDefine(
		"HOST_LOCK_NAME_TOO_LONG",
		"Actual lock name exceeds the limit of {maxBytes} bytes",
		gcode.CodeInvalidParameter,
	)
	// CodeHostLockLeaseTooShort reports that the requested lease is below the minimum.
	CodeHostLockLeaseTooShort = bizerr.MustDefine(
		"HOST_LOCK_LEASE_TOO_SHORT",
		"Lock lease cannot be shorter than {minLease}",
		gcode.CodeInvalidParameter,
	)
	// CodeHostLockLeaseTooLong reports that the requested lease is above the maximum.
	CodeHostLockLeaseTooLong = bizerr.MustDefine(
		"HOST_LOCK_LEASE_TOO_LONG",
		"Lock lease cannot be longer than {maxLease}",
		gcode.CodeInvalidParameter,
	)
	// CodeHostLockTicketMarshalFailed reports that a lock ticket cannot be serialized.
	CodeHostLockTicketMarshalFailed = bizerr.MustDefine(
		"HOST_LOCK_TICKET_MARSHAL_FAILED",
		"Failed to serialize lock ticket",
		gcode.CodeInternalError,
	)
	// CodeHostLockTicketRequired reports that a lock ticket is required.
	CodeHostLockTicketRequired = bizerr.MustDefine(
		"HOST_LOCK_TICKET_REQUIRED",
		"Lock ticket cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeHostLockTicketParseFailed reports that a lock ticket cannot be base64-decoded.
	CodeHostLockTicketParseFailed = bizerr.MustDefine(
		"HOST_LOCK_TICKET_PARSE_FAILED",
		"Failed to parse lock ticket",
		gcode.CodeInvalidParameter,
	)
	// CodeHostLockTicketUnmarshalFailed reports that a decoded lock ticket is not valid JSON.
	CodeHostLockTicketUnmarshalFailed = bizerr.MustDefine(
		"HOST_LOCK_TICKET_UNMARSHAL_FAILED",
		"Failed to deserialize lock ticket",
		gcode.CodeInvalidParameter,
	)
	// CodeHostLockTicketInvalid reports that a lock ticket has invalid required fields.
	CodeHostLockTicketInvalid = bizerr.MustDefine(
		"HOST_LOCK_TICKET_INVALID",
		"Lock ticket content is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeHostLockTicketPluginMismatch reports that a ticket belongs to a different plugin.
	CodeHostLockTicketPluginMismatch = bizerr.MustDefine(
		"HOST_LOCK_TICKET_PLUGIN_MISMATCH",
		"Lock ticket plugin identity does not match",
		gcode.CodeInvalidParameter,
	)
	// CodeHostLockTicketResourceMismatch reports that a ticket belongs to a different logical lock resource.
	CodeHostLockTicketResourceMismatch = bizerr.MustDefine(
		"HOST_LOCK_TICKET_RESOURCE_MISMATCH",
		"Lock ticket logical resource does not match",
		gcode.CodeInvalidParameter,
	)
)
