// This file defines authentication business error codes.

package auth

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeAuthInvalidCredentials reports invalid login credentials.
	// messageKey is derived as error.auth.invalid.credentials.
	CodeAuthInvalidCredentials = bizerr.MustDefine(
		"AUTH_INVALID_CREDENTIALS",
		"Invalid username or password",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthIPBlacklisted reports a login attempt from a denied IP address.
	// messageKey is derived as error.auth.ip.blacklisted.
	CodeAuthIPBlacklisted = bizerr.MustDefine(
		"AUTH_IP_BLACKLISTED",
		"Login IP is blacklisted",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthClientTypeInvalid reports a missing or unknown user-session client type.
	// messageKey is derived as error.auth.client.type.invalid.
	CodeAuthClientTypeInvalid = bizerr.MustDefine(
		"AUTH_CLIENT_TYPE_INVALID",
		"Client type is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeAuthUserDisabled reports a disabled user login attempt.
	// messageKey is derived as error.auth.user.disabled.
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
	// messageKey is derived as error.auth.login.state.update.failed.
	CodeAuthLoginStateUpdateFailed = bizerr.MustDefine(
		"AUTH_LOGIN_STATE_UPDATE_FAILED",
		"Failed to update last login time",
		gcode.CodeInternalError,
	)
	// CodeAuthExternalIdentityInvalid reports that an external login request
	// carries an empty provider or subject and cannot be resolved to a stable
	// external identity key.
	// messageKey is derived as error.auth.external.identity.invalid.
	CodeAuthExternalIdentityInvalid = bizerr.MustDefine(
		"AUTH_EXTERNAL_IDENTITY_INVALID",
		"External authentication provider returned an invalid identity",
		gcode.CodeInvalidParameter,
	)
	// CodeAuthExternalUserNotProvisioned reports that a verified external
	// identity has no linked local account. The message is intentionally
	// uniform regardless of whether the captured email exists as another
	// account so external login never leaks account existence.
	// messageKey is derived as error.auth.external.user.not.provisioned.
	CodeAuthExternalUserNotProvisioned = bizerr.MustDefine(
		"AUTH_EXTERNAL_USER_NOT_PROVISIONED",
		"No local account is linked to this external identity",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthRegisterDisabled reports that public self-registration is turned off.
	// messageKey is derived as error.auth.register.disabled.
	CodeAuthRegisterDisabled = bizerr.MustDefine(
		"AUTH_REGISTER_DISABLED",
		"Public registration is disabled",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthForgetPasswordDisabled reports that password recovery is turned off.
	// messageKey is derived as error.auth.forget.password.disabled.
	CodeAuthForgetPasswordDisabled = bizerr.MustDefine(
		"AUTH_FORGET_PASSWORD_DISABLED",
		"Password recovery is disabled",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthPasswordResetUnavailable reports that recovery cannot send email.
	// messageKey is derived as error.auth.password.reset.unavailable.
	CodeAuthPasswordResetUnavailable = bizerr.MustDefine(
		"AUTH_PASSWORD_RESET_UNAVAILABLE",
		"Password recovery is temporarily unavailable",
		gcode.CodeOperationFailed,
	)
	// CodeAuthPasswordResetTokenInvalid reports an invalid or used reset token.
	// messageKey is derived as error.auth.password.reset.token.invalid.
	CodeAuthPasswordResetTokenInvalid = bizerr.MustDefine(
		"AUTH_PASSWORD_RESET_TOKEN_INVALID",
		"Password reset token is invalid or expired",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthRegisterUsernameExists reports that the registration username is already used.
	// messageKey is derived as error.auth.register.username.exists.
	CodeAuthRegisterUsernameExists = bizerr.MustDefine(
		"AUTH_REGISTER_USERNAME_EXISTS",
		"Username already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeAuthRegisterEmailExists reports that the registration email is already used.
	// messageKey is derived as error.auth.register.email.exists.
	CodeAuthRegisterEmailExists = bizerr.MustDefine(
		"AUTH_REGISTER_EMAIL_EXISTS",
		"Email is already registered",
		gcode.CodeInvalidParameter,
	)
	// CodeAuthRegisterRateLimited reports that registration is rate limited.
	// messageKey is derived as error.auth.register.rate.limited.
	CodeAuthRegisterRateLimited = bizerr.MustDefine(
		"AUTH_REGISTER_RATE_LIMITED",
		"Too many registration attempts, please try again later",
		gcode.CodeOperationFailed,
	)
	// CodeAuthForgetPasswordRateLimited reports that recovery is rate limited.
	// messageKey is derived as error.auth.forget.password.rate.limited.
	CodeAuthForgetPasswordRateLimited = bizerr.MustDefine(
		"AUTH_FORGET_PASSWORD_RATE_LIMITED",
		"Too many password recovery attempts, please try again later",
		gcode.CodeOperationFailed,
	)
	// CodeAuthDefaultRoleMissing reports that the built-in standard user role is missing.
	// messageKey is derived as error.auth.default.role.missing.
	CodeAuthDefaultRoleMissing = bizerr.MustDefine(
		"AUTH_DEFAULT_ROLE_MISSING",
		"Default user role is not available",
		gcode.CodeInternalError,
	)
)
