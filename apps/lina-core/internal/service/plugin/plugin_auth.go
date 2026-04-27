// This file exposes auth-event hook dispatch methods on the root plugin facade.

package plugin

import (
	"context"

	"lina-core/pkg/pluginhost"
)

// English fallback messages published with host authentication lifecycle events.
const (
	// AuthEventMessageLoginSuccessful is the English fallback for successful login messages.
	AuthEventMessageLoginSuccessful = "Login successful"
	// AuthEventMessageLoginFailed is the English fallback for generic failed login messages.
	AuthEventMessageLoginFailed = "Login failed"
	// AuthEventMessageLogoutSuccessful is the English fallback for successful logout messages.
	AuthEventMessageLogoutSuccessful = "Logout successful"
	// AuthEventMessageInvalidCredentials is the English fallback for invalid credential messages.
	AuthEventMessageInvalidCredentials = "Invalid username or password"
	// AuthEventMessageUserDisabled is the English fallback for disabled account messages.
	AuthEventMessageUserDisabled = "User account is disabled"
	// AuthEventMessageIPBlacklisted is the English fallback for blocked login IP messages.
	AuthEventMessageIPBlacklisted = "Login IP is blacklisted"
)

// HandleAuthLoginSucceeded dispatches a login-succeeded hook to all enabled plugins.
func (s *serviceImpl) HandleAuthLoginSucceeded(ctx context.Context, input AuthLoginSucceededInput) error {
	return s.dispatchAuthHookEvent(
		ctx,
		pluginhost.ExtensionPointAuthLoginSucceeded,
		input,
		pluginhost.AuthHookReasonLoginSuccessful,
		AuthEventMessageLoginSuccessful,
	)
}

// HandleAuthLoginFailed dispatches a login-failed hook to all enabled plugins.
func (s *serviceImpl) HandleAuthLoginFailed(ctx context.Context, input AuthLoginSucceededInput) error {
	return s.dispatchAuthHookEvent(
		ctx,
		pluginhost.ExtensionPointAuthLoginFailed,
		input,
		pluginhost.AuthHookReasonLoginFailed,
		AuthEventMessageLoginFailed,
	)
}

// HandleAuthLogoutSucceeded dispatches a logout-succeeded hook to all enabled plugins.
func (s *serviceImpl) HandleAuthLogoutSucceeded(ctx context.Context, input AuthLoginSucceededInput) error {
	return s.dispatchAuthHookEvent(
		ctx,
		pluginhost.ExtensionPointAuthLogoutSucceeded,
		input,
		pluginhost.AuthHookReasonLogoutSuccessful,
		AuthEventMessageLogoutSuccessful,
	)
}

// dispatchAuthHookEvent normalizes common auth payload defaults before
// forwarding the event to the shared integration hook dispatcher.
func (s *serviceImpl) dispatchAuthHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	input AuthLoginSucceededInput,
	defaultReason string,
	defaultMessage string,
) error {
	if input.ClientType == "" {
		input.ClientType = "web"
	}
	if input.Reason == "" {
		input.Reason = defaultReason
	}
	if input.Message == "" {
		input.Message = defaultMessage
	}
	return s.integrationSvc.DispatchPluginHookEvent(
		ctx,
		event,
		pluginhost.BuildAuthHookPayloadValues(pluginhost.AuthHookPayloadInput{
			UserName:   input.UserName,
			Status:     input.Status,
			IP:         input.Ip,
			ClientType: input.ClientType,
			Browser:    input.Browser,
			OS:         input.Os,
			Message:    input.Message,
			Reason:     input.Reason,
		}),
	)
}
