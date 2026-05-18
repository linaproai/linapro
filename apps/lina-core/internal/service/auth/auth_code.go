// This file defines authentication business error codes.

package auth

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeAuthInvalidCredentials reports invalid login credentials.
	CodeAuthInvalidCredentials = bizerr.MustDefine(
		"AUTH_INVALID_CREDENTIALS",
		"Invalid username or password",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthIPBlacklisted reports a login attempt from a denied IP address.
	CodeAuthIPBlacklisted = bizerr.MustDefineWithKey(
		"AUTH_IP_BLACKLISTED",
		"error.auth.login.ipBlacklisted",
		"Login IP is blacklisted",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthUserDisabled reports a disabled user login attempt.
	CodeAuthUserDisabled = bizerr.MustDefine(
		"AUTH_USER_DISABLED",
		"User account is disabled",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthTokenInvalid reports an invalid or revoked JWT.
	CodeAuthTokenInvalid = bizerr.MustDefine(
		"AUTH_TOKEN_INVALID",
		"Authentication token is invalid",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthPreTokenInvalid reports an invalid, expired, or already used pre-login token.
	CodeAuthPreTokenInvalid = bizerr.MustDefine(
		"AUTH_PRE_TOKEN_INVALID",
		"Pre-login token is invalid or expired",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthTokenStateUnavailable reports that shared auth token state cannot be read or written.
	CodeAuthTokenStateUnavailable = bizerr.MustDefine(
		"AUTH_TOKEN_STATE_UNAVAILABLE",
		"Authentication token state is temporarily unavailable",
		gcode.CodeInternalError,
	)
	// CodeAuthTenantUnavailable reports that a tenant-bound user has no active tenant to sign in to.
	CodeAuthTenantUnavailable = bizerr.MustDefine(
		"AUTH_TENANT_UNAVAILABLE",
		"Tenant is not available",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthLoginStateUpdateFailed reports that login succeeded but last-login state cannot be persisted.
	CodeAuthLoginStateUpdateFailed = bizerr.MustDefineWithKey(
		"AUTH_LOGIN_STATE_UPDATE_FAILED",
		"error.auth.login.updateLastLoginFailed",
		"Failed to update last login time",
		gcode.CodeInternalError,
	)
)
