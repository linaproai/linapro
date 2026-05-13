// Package operlog implements the monitor-operlog plugin HTTP controllers.
package operlog

import (
	operlogapi "lina-plugin-monitor-operlog/backend/api/operlog"
	v1 "lina-plugin-monitor-operlog/backend/api/operlog/v1"
	"lina-plugin-monitor-operlog/backend/internal/model/operlogtype"
	operlogsvc "lina-plugin-monitor-operlog/backend/internal/service/operlog"
)

// ControllerV1 is the operation-log controller.
type ControllerV1 struct {
	operLogSvc operlogsvc.Service // operation-log service
}

// NewV1 creates and returns a new monitor-operlog controller instance.
func NewV1(operLogSvc operlogsvc.Service) operlogapi.IOperlogV1 {
	return &ControllerV1{operLogSvc: operLogSvc}
}

// toAPIOperLogEntity converts one service-layer operation-log entity into the API DTO projection.
func toAPIOperLogEntity(entity *operlogsvc.OperLogEntity) *v1.OperLogEntity {
	if entity == nil {
		return nil
	}
	return &v1.OperLogEntity{
		Id:                 entity.Id,
		TenantId:           entity.TenantId,
		ActingUserId:       entity.ActingUserId,
		OnBehalfOfTenantId: entity.OnBehalfOfTenantId,
		IsImpersonation:    entity.IsImpersonation,
		Title:              entity.Title,
		OperSummary:        entity.OperSummary,
		OperType:           operlogtype.Normalize(entity.OperType),
		Method:             entity.Method,
		RequestMethod:      entity.RequestMethod,
		OperName:           entity.OperName,
		OperUrl:            entity.OperUrl,
		OperIp:             entity.OperIp,
		OperParam:          entity.OperParam,
		JsonResult:         entity.JsonResult,
		Status:             entity.Status,
		ErrorMsg:           entity.ErrorMsg,
		CostTime:           entity.CostTime,
		OperTime:           entity.OperTime,
	}
}

// normalizeOperTypePointer converts an optional request value into the shared
// semantic operation type used by operation-log records.
func normalizeOperTypePointer(value *string) *operlogtype.OperType {
	if value == nil {
		return nil
	}
	operType := operlogtype.Normalize(*value)
	if !operlogtype.IsSupported(operType) {
		return nil
	}
	return &operType
}
