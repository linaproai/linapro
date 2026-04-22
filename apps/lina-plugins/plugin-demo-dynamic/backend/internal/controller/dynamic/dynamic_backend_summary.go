// Backend summary route controller.

package dynamic

import (
	"context"
	"strings"

	"lina-core/pkg/pluginbridge"
	"lina-plugin-demo-dynamic/backend/api/dynamic/v1"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// BackendSummary returns plugin bridge execution summary including plugin
// identity, route metadata, and current user context.
func (c *Controller) BackendSummary(
	ctx context.Context,
	_ *v1.BackendSummaryReq,
) (res *v1.BackendSummaryRes, err error) {
	payload := c.dynamicSvc.BuildBackendSummaryPayload(
		buildBackendSummaryRouteInput(pluginbridge.RequestEnvelopeFromContext(ctx)),
	)
	if err = pluginbridge.SetResponseHeader(ctx, "X-Lina-Plugin-Bridge", "plugin-demo-dynamic"); err != nil {
		return nil, err
	}
	if err = pluginbridge.SetResponseHeader(ctx, "X-Lina-Plugin-Middleware", "backend-summary"); err != nil {
		return nil, err
	}
	return &v1.BackendSummaryRes{
		Message:       payload.Message,
		PluginID:      payload.PluginID,
		PublicPath:    payload.PublicPath,
		Access:        payload.Access,
		Permission:    payload.Permission,
		Authenticated: payload.Authenticated,
		Username:      payload.Username,
		IsSuperAdmin:  payload.IsSuperAdmin,
	}, nil
}

// buildBackendSummaryRouteInput extracts backend summary route metadata and
// identity context from the bridge request.
func buildBackendSummaryRouteInput(request *pluginbridge.BridgeRequestEnvelopeV1) *dynamicservice.BackendSummaryInput {
	input := &dynamicservice.BackendSummaryInput{}
	if request == nil {
		return input
	}

	input.PluginID = strings.TrimSpace(request.PluginID)
	if request.Route != nil {
		input.PublicPath = strings.TrimSpace(request.Route.PublicPath)
		input.Access = strings.TrimSpace(request.Route.Access)
		input.Permission = strings.TrimSpace(request.Route.Permission)
	}
	if request.Identity != nil {
		input.Authenticated = request.Identity.UserID > 0
		input.HasIdentity = true
		input.Username = strings.TrimSpace(request.Identity.Username)
		input.IsSuperAdmin = request.Identity.IsSuperAdmin
	}
	return input
}
