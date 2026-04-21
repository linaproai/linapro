// This file exposes request-audit hook dispatch methods on the root plugin facade.

package plugin

import (
	"context"

	"lina-core/pkg/pluginhost"
)

// HandleAuditRecorded dispatches one request-audit hook to all enabled plugins.
func (s *serviceImpl) HandleAuditRecorded(ctx context.Context, input AuditRecordedInput) error {
	return s.integrationSvc.DispatchPluginHookEvent(
		ctx,
		pluginhost.ExtensionPointAuditRecorded,
		pluginhost.BuildAuditHookPayloadValues(pluginhost.AuditHookPayloadInput{
			Title:         input.Title,
			OperSummary:   input.OperSummary,
			OperType:      input.OperType,
			Method:        input.Method,
			RequestMethod: input.RequestMethod,
			OperName:      input.OperName,
			OperURL:       input.OperUrl,
			OperIP:        input.OperIp,
			OperParam:     input.OperParam,
			JSONResult:    input.JsonResult,
			Status:        input.Status,
			ErrorMsg:      input.ErrorMsg,
			CostTime:      input.CostTime,
		}),
	)
}
