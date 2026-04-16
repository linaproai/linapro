// This file exposes auth-event hook dispatch methods on the root plugin facade.

package plugin

import (
	"context"

	"lina-core/pkg/pluginhost"
)

// HandleAuthLoginSucceeded dispatches a login-succeeded hook to all enabled plugins.
func (s *serviceImpl) HandleAuthLoginSucceeded(ctx context.Context, input AuthLoginSucceededInput) error {
	return s.dispatchAuthHookEvent(ctx, pluginhost.ExtensionPointAuthLoginSucceeded, input, "登录成功")
}

// HandleAuthLoginFailed dispatches a login-failed hook to all enabled plugins.
func (s *serviceImpl) HandleAuthLoginFailed(ctx context.Context, input AuthLoginSucceededInput) error {
	return s.dispatchAuthHookEvent(ctx, pluginhost.ExtensionPointAuthLoginFailed, input, "登录失败")
}

// HandleAuthLogoutSucceeded dispatches a logout-succeeded hook to all enabled plugins.
func (s *serviceImpl) HandleAuthLogoutSucceeded(ctx context.Context, input AuthLoginSucceededInput) error {
	return s.dispatchAuthHookEvent(ctx, pluginhost.ExtensionPointAuthLogoutSucceeded, input, "登出成功")
}

func (s *serviceImpl) dispatchAuthHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	input AuthLoginSucceededInput,
	defaultMessage string,
) error {
	if input.ClientType == "" {
		input.ClientType = "web"
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
		}),
	)
}
