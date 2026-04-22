// Package operlog implements the monitor-operlog plugin HTTP controllers.
package operlog

import (
	operlogapi "lina-plugin-monitor-operlog/backend/api/operlog"
	v1 "lina-plugin-monitor-operlog/backend/api/operlog/v1"
	operlogsvc "lina-plugin-monitor-operlog/backend/internal/service/operlog"
)

// ControllerV1 is the operation-log controller.
type ControllerV1 struct {
	operLogSvc operlogsvc.Service // operation-log service
}

// NewV1 creates and returns a new monitor-operlog controller instance.
func NewV1() operlogapi.IOperlogV1 {
	return &ControllerV1{operLogSvc: operlogsvc.New()}
}

// toAPIOperLogEntity converts one service-layer operation-log entity into the API DTO projection.
func toAPIOperLogEntity(entity *operlogsvc.OperLogEntity) *v1.OperLogEntity {
	if entity == nil {
		return nil
	}
	return &v1.OperLogEntity{
		Id:            entity.Id,
		Title:         entity.Title,
		OperSummary:   entity.OperSummary,
		OperType:      entity.OperType,
		Method:        entity.Method,
		RequestMethod: entity.RequestMethod,
		OperName:      entity.OperName,
		OperUrl:       entity.OperUrl,
		OperIp:        entity.OperIp,
		OperParam:     entity.OperParam,
		JsonResult:    entity.JsonResult,
		Status:        entity.Status,
		ErrorMsg:      entity.ErrorMsg,
		CostTime:      entity.CostTime,
		OperTime:      entity.OperTime,
	}
}
