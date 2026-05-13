// Package loginlog implements the monitor-loginlog plugin HTTP controllers.
package loginlog

import (
	loginlogapi "lina-plugin-monitor-loginlog/backend/api/loginlog"
	v1 "lina-plugin-monitor-loginlog/backend/api/loginlog/v1"
	loginlogsvc "lina-plugin-monitor-loginlog/backend/internal/service/loginlog"
)

// ControllerV1 is the login-log controller.
type ControllerV1 struct {
	loginLogSvc loginlogsvc.Service // login-log service
}

// NewV1 creates and returns a new monitor-loginlog controller instance.
func NewV1(loginLogSvc loginlogsvc.Service) loginlogapi.ILoginlogV1 {
	return &ControllerV1{loginLogSvc: loginLogSvc}
}

// toAPILoginLogEntity converts one service-layer login-log entity into the API DTO projection.
func toAPILoginLogEntity(entity *loginlogsvc.LoginLogEntity) *v1.LoginLogEntity {
	if entity == nil {
		return nil
	}
	return &v1.LoginLogEntity{
		Id:                 entity.Id,
		TenantId:           entity.TenantId,
		ActingUserId:       entity.ActingUserId,
		OnBehalfOfTenantId: entity.OnBehalfOfTenantId,
		IsImpersonation:    entity.IsImpersonation,
		UserName:           entity.UserName,
		Status:             entity.Status,
		Ip:                 entity.Ip,
		Browser:            entity.Browser,
		Os:                 entity.Os,
		Msg:                entity.Msg,
		LoginTime:          entity.LoginTime,
	}
}
